package phonesafe

import (
	"slices"
	"sort"

	"github.com/shepard-labs/go-phonesafe/internal/metadata"
)

// --- Internal helpers ---

// getDescForType returns the PhoneNumberDesc for a given type from the metadata.
// For TypeFixedLineOrMobile, returns the FixedLine descriptor (caller merges Mobile).
// For TypeUnknown, returns GeneralDesc.
func getDescForType(meta *metadata.PhoneMetadata, numType PhoneNumberType) metadata.PhoneNumberDesc {
	switch numType {
	case TypePremiumRate:
		return meta.PremiumRate
	case TypeTollFree:
		return meta.TollFree
	case TypeMobile:
		return meta.Mobile
	case TypeFixedLine, TypeFixedLineOrMobile:
		return meta.FixedLine
	case TypeSharedCost:
		return meta.SharedCost
	case TypeVOIP:
		return meta.VOIP
	case TypePersonalNumber:
		return meta.PersonalNumber
	case TypePager:
		return meta.Pager
	case TypeUAN:
		return meta.UAN
	case TypeVoicemail:
		return meta.Voicemail
	default:
		return meta.GeneralDesc
	}
}

// descHasPossibleNumberData returns true if the descriptor has meaningful
// possible length data (not just [-1] indicating unsupported type).
func descHasPossibleNumberData(desc metadata.PhoneNumberDesc) bool {
	return len(desc.PossibleLengths) != 1 || desc.PossibleLengths[0] != -1
}

// aggregateTypeLengths collects all possible lengths from all type descriptors
// in the metadata. Used when GeneralDesc.PossibleLengths is empty (our codegen
// doesn't pre-compute the merged set like upstream does at build time).
func aggregateTypeLengths(meta *metadata.PhoneMetadata) (possibleLengths, localLengths []int32) {
	descs := []metadata.PhoneNumberDesc{
		meta.FixedLine, meta.Mobile, meta.TollFree, meta.PremiumRate,
		meta.SharedCost, meta.PersonalNumber, meta.VOIP, meta.Pager,
		meta.UAN, meta.Voicemail,
	}

	var allLengths, allLocal []int32
	for _, d := range descs {
		if len(d.PossibleLengths) > 0 && d.PossibleLengths[0] != -1 {
			allLengths = append(allLengths, d.PossibleLengths...)
		}
		if len(d.PossibleLengthsLocal) > 0 {
			allLocal = append(allLocal, d.PossibleLengthsLocal...)
		}
	}

	if len(allLengths) > 0 {
		slices.Sort(allLengths)
		possibleLengths = slices.Compact(allLengths)
	}
	if len(allLocal) > 0 {
		slices.Sort(allLocal)
		localLengths = slices.Compact(allLocal)
	}
	return
}

// testNumberLengthForType checks the length of a national number against the
// possible lengths for a specific phone number type. Handles the
// TypeFixedLineOrMobile case by merging lengths from both descriptors.
func (u *PhoneNumberUtil) testNumberLengthForType(number string, meta *metadata.PhoneMetadata, numType PhoneNumberType) ValidationResult {
	desc := getDescForType(meta, numType)

	// When a sub-description has no possible lengths, it inherits from GeneralDesc.
	possibleLengths := desc.PossibleLengths
	if len(possibleLengths) == 0 {
		possibleLengths = meta.GeneralDesc.PossibleLengths
	}
	localLengths := desc.PossibleLengthsLocal

	if numType == TypeFixedLineOrMobile {
		fixedLineDesc := getDescForType(meta, TypeFixedLine)
		if !descHasPossibleNumberData(fixedLineDesc) {
			// No fixed-line data — just check mobile.
			return u.testNumberLengthForType(number, meta, TypeMobile)
		}

		mobileDesc := getDescForType(meta, TypeMobile)
		if descHasPossibleNumberData(mobileDesc) {
			// Merge mobile lengths into the fixed-line lengths.
			mobileLengths := mobileDesc.PossibleLengths
			if len(mobileLengths) == 0 {
				mobileLengths = meta.GeneralDesc.PossibleLengths
			}

			// Create merged + sorted + deduped slice.
			merged := make([]int32, 0, len(possibleLengths)+len(mobileLengths))
			merged = append(merged, possibleLengths...)
			merged = append(merged, mobileLengths...)
			slices.Sort(merged)
			possibleLengths = slices.Compact(merged)

			// Merge local lengths too.
			if len(localLengths) == 0 {
				localLengths = mobileDesc.PossibleLengthsLocal
			} else if len(mobileDesc.PossibleLengthsLocal) > 0 {
				mergedLocal := make([]int32, 0, len(localLengths)+len(mobileDesc.PossibleLengthsLocal))
				mergedLocal = append(mergedLocal, localLengths...)
				mergedLocal = append(mergedLocal, mobileDesc.PossibleLengthsLocal...)
				slices.Sort(mergedLocal)
				localLengths = slices.Compact(mergedLocal)
			}
		}
	}

	// If GeneralDesc has no lengths and we're checking TypeUnknown, aggregate
	// from all available type descriptors (mirrors upstream build-time merge).
	if len(possibleLengths) == 0 && numType == TypeUnknown {
		possibleLengths, localLengths = aggregateTypeLengths(meta)
	}

	// No length info — fall back to NSN range check.
	if len(possibleLengths) == 0 {
		l := int32(len(number))
		if l < minLengthForNSN {
			return TooShort
		}
		if l > maxLengthForNSN {
			return TooLong
		}
		return IsPossible
	}

	// -1 in possible lengths means the type is not supported at all.
	if possibleLengths[0] == -1 {
		return InvalidLength
	}

	actualLength := int32(len(number))

	// Check local-only lengths.
	if slices.Contains(localLengths, actualLength) {
		return IsPossibleLocalOnly
	}

	// Check possible lengths.
	minLength := possibleLengths[0]
	if minLength == actualLength {
		return IsPossible
	}
	if minLength > actualLength {
		return TooShort
	}
	maxLength := possibleLengths[len(possibleLengths)-1]
	if maxLength < actualLength {
		return TooLong
	}

	// Check remaining lengths.
	if slices.Contains(possibleLengths[1:], actualLength) {
		return IsPossible
	}
	return InvalidLength
}

// isNumberMatchingDesc checks if a national number matches a PhoneNumberDesc
// (length check + pattern match).
func (u *PhoneNumberUtil) isNumberMatchingDesc(nationalNumber string, desc metadata.PhoneNumberDesc) bool {
	// Quick reject by possible lengths (if present).
	if len(desc.PossibleLengths) > 0 {
		actualLength := int32(len(nationalNumber))
		if !slices.Contains(desc.PossibleLengths, actualLength) {
			return false
		}
	}

	// Full pattern match.
	pattern := desc.NationalNumberPattern
	if pattern == "" {
		return false
	}
	return u.matchesEntirely(pattern, nationalNumber)
}

// getRegionCodeForNumber returns the region code for a parsed phone number.
// When a country code maps to multiple regions, disambiguates using leadingDigits
// or pattern matching.
func (u *PhoneNumberUtil) getRegionCodeForNumber(number PhoneNumber) string {
	cc := number.CountryCode
	regions := metadata.CountryCodeToRegions[cc]
	if len(regions) == 0 {
		// Check if it's a non-geographical entity.
		if _, ok := metadata.NonGeoMetadata[cc]; ok {
			return regionCodeNonGeo
		}
		return "ZZ"
	}
	if len(regions) == 1 {
		return regions[0]
	}

	// Multiple regions share this CC — disambiguate.
	nationalNumber := GetNationalSignificantNumber(number)
	for _, region := range regions {
		meta := u.getMetadataForRegion(region)
		if meta == nil {
			continue
		}
		if meta.LeadingDigits != "" {
			// Use leadingDigits prefix match.
			re, err := u.rc.getOrError(`^(?:` + meta.LeadingDigits + `)`)
			if err == nil && re.MatchString(nationalNumber) {
				return region
			}
		} else if u.getNumberTypeHelper(nationalNumber, meta) != TypeUnknown {
			return region
		}
	}
	return "ZZ"
}

// getCountryCodeForValidRegion returns the country calling code for a valid region.
func (u *PhoneNumberUtil) getCountryCodeForValidRegion(region string) int32 {
	meta := u.getMetadataForRegion(region)
	if meta == nil {
		return 0
	}
	return meta.CountryCode
}

// --- Public API ---

// IsPossibleNumber performs a fast length-only check (no pattern matching).
func (u *PhoneNumberUtil) IsPossibleNumber(number PhoneNumber) bool {
	result := u.IsPossibleNumberWithReason(number)
	return result == IsPossible || result == IsPossibleLocalOnly
}

// IsPossibleNumberWithReason returns the specific reason a number is/isn't possible.
func (u *PhoneNumberUtil) IsPossibleNumberWithReason(number PhoneNumber) ValidationResult {
	return u.IsPossibleNumberForTypeWithReason(number, TypeUnknown)
}

// IsPossibleNumberForType checks possibility for a specific number type.
func (u *PhoneNumberUtil) IsPossibleNumberForType(number PhoneNumber, numType PhoneNumberType) bool {
	result := u.IsPossibleNumberForTypeWithReason(number, numType)
	return result == IsPossible || result == IsPossibleLocalOnly
}

// IsPossibleNumberForTypeWithReason returns the reason with type context.
func (u *PhoneNumberUtil) IsPossibleNumberForTypeWithReason(number PhoneNumber, numType PhoneNumberType) ValidationResult {
	nationalNumber := GetNationalSignificantNumber(number)
	cc := number.CountryCode

	// Note: For regions that share a country calling code (e.g. NANPA), we use
	// rules from the default region since getRegionCodeForNumber won't work for
	// possible-but-invalid numbers.
	if !u.hasValidCountryCallingCode(cc) {
		return InvalidCountryCode
	}

	region := u.getRegionCodeForCountryCode(cc)
	meta := u.getMetadataForRegionOrCallingCode(cc, region)
	return u.testNumberLengthForType(nationalNumber, meta, numType)
}

// IsValidNumber performs full validation: length + leading digit pattern matching.
// Does NOT verify the number is currently assigned/in-use.
func (u *PhoneNumberUtil) IsValidNumber(number PhoneNumber) bool {
	region := u.getRegionCodeForNumber(number)
	return u.IsValidNumberForRegion(number, region)
}

// IsValidNumberForRegion checks if a number is valid specifically for the given region.
func (u *PhoneNumberUtil) IsValidNumberForRegion(number PhoneNumber, region string) bool {
	cc := number.CountryCode
	meta := u.getMetadataForRegionOrCallingCode(cc, region)
	if meta == nil {
		return false
	}

	// If region is a real region (not "001"), verify the CC matches.
	if region != regionCodeNonGeo {
		regionCC := u.getCountryCodeForValidRegion(region)
		if regionCC == 0 || cc != regionCC {
			return false
		}
	}

	nationalNumber := GetNationalSignificantNumber(number)
	return u.getNumberTypeHelper(nationalNumber, meta) != TypeUnknown
}

// GetRegionCodeForNumber returns the CLDR region code for the given phone number.
// Returns "ZZ" for unknown/invalid numbers.
func (u *PhoneNumberUtil) GetRegionCodeForNumber(number PhoneNumber) string {
	return u.getRegionCodeForNumber(number)
}

// CanBeInternationallyDialled returns true if the number can be dialed from abroad.
func (u *PhoneNumberUtil) CanBeInternationallyDialled(number PhoneNumber) bool {
	meta := u.getMetadataForRegionOrCallingCode(number.CountryCode,
		u.getRegionCodeForNumber(number))
	if meta == nil {
		return true
	}
	nationalNumber := GetNationalSignificantNumber(number)
	// If no NoInternationalDialling pattern, everything is diallable.
	if meta.NoInternationalDialling.NationalNumberPattern == "" {
		return true
	}
	return !u.isNumberMatchingDesc(nationalNumber, meta.NoInternationalDialling)
}

// IsAlphaNumber returns true if the number contains 3+ alpha characters,
// making it a valid vanity number like "1-800-MICROSOFT".
func (u *PhoneNumberUtil) IsAlphaNumber(number string) bool {
	if !u.isViablePhoneNumber(number) {
		return false
	}
	// Strip extension before checking.
	stripped, _ := u.maybeStripExtension(number)
	// Check if the remaining portion has enough alpha characters.
	return validAlphaPattern.MatchString(stripped)
}

// GetRegionCodesForCountryCode returns all regions sharing the given country calling code.
func (u *PhoneNumberUtil) GetRegionCodesForCountryCode(cc int32) []string {
	regions := metadata.CountryCodeToRegions[cc]
	if len(regions) == 0 {
		return nil
	}
	result := make([]string, len(regions))
	copy(result, regions)
	return result
}

// GetCountryCodeForRegion returns the country calling code for a valid region.
// Returns 0 for invalid regions.
func (u *PhoneNumberUtil) GetCountryCodeForRegion(region string) int32 {
	if !u.isValidRegionCode(region) {
		return 0
	}
	return u.getCountryCodeForValidRegion(region)
}

// GetRegionCodeForCountryCode returns the main region for a calling code.
// For shared calling codes (e.g., +1), returns the main country (e.g., "US").
// Returns empty string if the calling code is unknown.
func (u *PhoneNumberUtil) GetRegionCodeForCountryCode(cc int) string {
	regions := metadata.CountryCodeToRegions[int32(cc)]
	if len(regions) == 0 {
		return ""
	}
	return regions[0]
}

// GetSupportedRegions returns all supported CLDR region codes (excluding "001").
// The returned slice is sorted alphabetically.
func (u *PhoneNumberUtil) GetSupportedRegions() []string {
	regions := make([]string, 0, len(metadata.RegionMetadata))
	for r := range metadata.RegionMetadata {
		if r != "001" {
			regions = append(regions, r)
		}
	}
	sort.Strings(regions)
	return regions
}

// GetSupportedCallingCodes returns all supported country calling codes.
// The returned slice is sorted numerically.
func (u *PhoneNumberUtil) GetSupportedCallingCodes() []int {
	seen := make(map[int32]struct{}, len(metadata.CountryCodeToRegions)+len(metadata.NonGeoMetadata))
	for cc := range metadata.CountryCodeToRegions {
		seen[cc] = struct{}{}
	}
	for cc := range metadata.NonGeoMetadata {
		seen[cc] = struct{}{}
	}
	codes := make([]int, 0, len(seen))
	for cc := range seen {
		codes = append(codes, int(cc))
	}
	sort.Ints(codes)
	return codes
}

// GetSupportedTypesForRegion returns the phone number types that have
// valid patterns defined for the given region.
func (u *PhoneNumberUtil) GetSupportedTypesForRegion(region string) []PhoneNumberType {
	if !u.isValidRegionCode(region) {
		return nil
	}
	meta := u.getMetadataForRegion(region)
	if meta == nil {
		return nil
	}
	var types []PhoneNumberType
	for _, t := range []struct {
		typ  PhoneNumberType
		desc metadata.PhoneNumberDesc
	}{
		{TypeFixedLine, meta.FixedLine},
		{TypeMobile, meta.Mobile},
		{TypeTollFree, meta.TollFree},
		{TypePremiumRate, meta.PremiumRate},
		{TypeSharedCost, meta.SharedCost},
		{TypePersonalNumber, meta.PersonalNumber},
		{TypeVOIP, meta.VOIP},
		{TypePager, meta.Pager},
		{TypeUAN, meta.UAN},
		{TypeVoicemail, meta.Voicemail},
	} {
		if t.desc.NationalNumberPattern != "" {
			types = append(types, t.typ)
		}
	}
	return types
}

// IsMobileNumberPortableRegion returns true if the region supports
// mobile number portability.
func (u *PhoneNumberUtil) IsMobileNumberPortableRegion(region string) bool {
	meta := u.getMetadataForRegion(region)
	if meta == nil {
		return false
	}
	return meta.MobileNumberPortable
}
