package phonesafe

import (
	"slices"

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

// getNumberTypeHelper determines the PhoneNumberType by testing the national
// number against each type descriptor in the metadata.
func (u *PhoneNumberUtil) getNumberTypeHelper(nationalNumber string, meta *metadata.PhoneMetadata) PhoneNumberType {
	// First check general desc — if it doesn't match, number is unknown.
	if !u.isNumberMatchingDesc(nationalNumber, meta.GeneralDesc) {
		return TypeUnknown
	}

	if u.isNumberMatchingDesc(nationalNumber, meta.PremiumRate) {
		return TypePremiumRate
	}
	if u.isNumberMatchingDesc(nationalNumber, meta.TollFree) {
		return TypeTollFree
	}
	if u.isNumberMatchingDesc(nationalNumber, meta.SharedCost) {
		return TypeSharedCost
	}
	if u.isNumberMatchingDesc(nationalNumber, meta.VOIP) {
		return TypeVOIP
	}
	if u.isNumberMatchingDesc(nationalNumber, meta.PersonalNumber) {
		return TypePersonalNumber
	}
	if u.isNumberMatchingDesc(nationalNumber, meta.Pager) {
		return TypePager
	}
	if u.isNumberMatchingDesc(nationalNumber, meta.UAN) {
		return TypeUAN
	}
	if u.isNumberMatchingDesc(nationalNumber, meta.Voicemail) {
		return TypeVoicemail
	}

	// Check fixed-line and mobile.
	isFixedLine := u.isNumberMatchingDesc(nationalNumber, meta.FixedLine)
	if isFixedLine {
		// If patterns are identical, it's FixedLineOrMobile.
		if meta.FixedLine.NationalNumberPattern == meta.Mobile.NationalNumberPattern {
			return TypeFixedLineOrMobile
		}
		if u.isNumberMatchingDesc(nationalNumber, meta.Mobile) {
			return TypeFixedLineOrMobile
		}
		return TypeFixedLine
	}

	// Check mobile only if patterns differ from fixed-line.
	if meta.FixedLine.NationalNumberPattern != meta.Mobile.NationalNumberPattern {
		if u.isNumberMatchingDesc(nationalNumber, meta.Mobile) {
			return TypeMobile
		}
	}

	return TypeUnknown
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

// GetNumberType returns the classification of a valid phone number.
// Returns TypeUnknown for invalid numbers.
func (u *PhoneNumberUtil) GetNumberType(number PhoneNumber) PhoneNumberType {
	region := u.getRegionCodeForNumber(number)
	meta := u.getMetadataForRegionOrCallingCode(number.CountryCode, region)
	if meta == nil {
		return TypeUnknown
	}
	nationalNumber := GetNationalSignificantNumber(number)
	return u.getNumberTypeHelper(nationalNumber, meta)
}

// GetRegionCodeForNumber returns the CLDR region code for the given phone number.
// Returns "ZZ" for unknown/invalid numbers.
func (u *PhoneNumberUtil) GetRegionCodeForNumber(number PhoneNumber) string {
	return u.getRegionCodeForNumber(number)
}

// IsNumberGeographical returns true if the number is associated with a
// geographic area (has an area code).
func (u *PhoneNumberUtil) IsNumberGeographical(number PhoneNumber) bool {
	numType := u.GetNumberType(number)
	return u.isNumberTypeGeographical(numType, number.CountryCode)
}

// isNumberTypeGeographical checks if a phone number type is geographical
// for the given country code.
func (u *PhoneNumberUtil) isNumberTypeGeographical(numType PhoneNumberType, cc int32) bool {
	return numType == TypeFixedLine || numType == TypeFixedLineOrMobile ||
		(numType == TypeMobile && isGeoMobileCountry(cc))
}

// geoMobileCountries lists country codes where mobile numbers are geographical.
// Matches upstream GEO_MOBILE_COUNTRIES set.
var geoMobileCountries = map[int32]bool{
	52: true, // Mexico
	54: true, // Argentina
	55: true, // Brazil
	62: true, // Indonesia
}

// isGeoMobileCountry returns true if the given country code has geographic mobile numbers.
func isGeoMobileCountry(cc int32) bool {
	return geoMobileCountries[cc]
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

