package phonesafe

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/shepard-labs/go-phonesafe/internal/metadata"
)

// Formatting constants matching upstream.
const (
	defaultExtnPrefix = " ext. "
)

// singleInternationalPrefixPattern matches a simple international prefix
// (digits, optionally followed by a tilde-separator and more digits).
var singleInternationalPrefixPattern = regexp.MustCompile(
	`^[\d]+(?:[~\x{2053}\x{223C}\x{FF5E}][\d]+)?$`)

// separatorPattern matches phone number separators (used in RFC3966 formatting).
var separatorPattern = regexp.MustCompile(`[-\x{2010}-\x{2015}\x{2212}\x{30FC}\x{FF0D}-\x{FF0F} \x{00A0}\x{00AD}\x{200B}\x{2060}\x{3000}()\x{FF08}\x{FF09}\x{FF3B}\x{FF3D}.\\/~\x{2053}\x{223C}]+`)

// leadingSeparatorPattern matches leading punctuation/separators to strip in RFC3966 formatting.
var leadingSeparatorPattern = regexp.MustCompile(`^[-\x{2010}-\x{2015}\x{2212}\x{30FC}\x{FF0D}-\x{FF0F} \x{00A0}\x{00AD}\x{200B}\x{2060}\x{3000}()\x{FF08}\x{FF09}\x{FF3B}\x{FF3D}.\\/~\x{2053}\x{223C}]+`)

// --- Internal formatting helpers ---

// prefixNumberWithCountryCallingCode prepends the country calling code
// with appropriate formatting based on the requested format.
func prefixNumberWithCountryCallingCode(cc int32, format PhoneNumberFormat, formattedNumber string) string {
	ccStr := strconv.FormatInt(int64(cc), 10)
	switch format {
	case FormatE164:
		return "+" + ccStr + formattedNumber
	case FormatInternational:
		return "+" + ccStr + " " + formattedNumber
	case FormatRFC3966:
		return "tel:+" + ccStr + "-" + formattedNumber
	default: // FormatNational
		return formattedNumber
	}
}

// maybeAppendFormattedExtension appends the extension in the appropriate format.
func maybeAppendFormattedExtension(number PhoneNumber, meta *metadata.PhoneMetadata, format PhoneNumberFormat, formattedNumber string) string {
	if number.Extension == "" {
		return formattedNumber
	}
	if format == FormatRFC3966 {
		return formattedNumber + rfc3966ExtnPrefix + number.Extension
	}
	if meta != nil && meta.PreferredExtnPrefix != "" {
		return formattedNumber + meta.PreferredExtnPrefix + number.Extension
	}
	return formattedNumber + defaultExtnPrefix + number.Extension
}

// chooseFormattingPatternForNumber selects the appropriate NumberFormat from
// available formats by checking leading digits and then full pattern match.
func (u *PhoneNumberUtil) chooseFormattingPatternForNumber(formats []metadata.NumberFormat, nationalNumber string) *metadata.NumberFormat {
	for i := range formats {
		f := &formats[i]
		size := len(f.LeadingDigitsPatterns)
		if size > 0 {
			// Use the last (most specific) leading digits pattern.
			lastPattern := f.LeadingDigitsPatterns[size-1]
			re, err := u.rc.getOrError(`^(?:` + lastPattern + `)`)
			if err != nil {
				continue
			}
			if !re.MatchString(nationalNumber) {
				continue
			}
		}
		// Check full pattern match.
		if u.matchesEntirely(f.Pattern, nationalNumber) {
			return f
		}
	}
	return nil
}

// resolveFormattingRule resolves $NP and $FG placeholders in a formatting rule.
func resolveFormattingRule(rule, nationalPrefix string) string {
	rule = strings.Replace(rule, "$NP", nationalPrefix, 1)
	rule = strings.Replace(rule, "$FG", "$1", 1)
	return rule
}

// formatNsnUsingPatternWithCarrier formats the national significant number
// using a specific pattern, optionally inserting a carrier code.
// nationalPrefix is needed to resolve $NP in formatting rules.
func (u *PhoneNumberUtil) formatNsnUsingPatternWithCarrier(nationalNumber string, pattern *metadata.NumberFormat, format PhoneNumberFormat, carrierCode string, nationalPrefix string) string {
	numberFormatRule := pattern.Format

	if format == FormatNational && carrierCode != "" && pattern.DomesticCarrierCodeFormattingRule != "" {
		// Resolve $NP, $CC, and $FG in the domestic carrier code rule.
		ccRule := resolveFormattingRule(pattern.DomesticCarrierCodeFormattingRule, nationalPrefix)
		ccRule = strings.Replace(ccRule, "$CC", carrierCode, 1)
		// The carrier code rule replaces the first $1 in the format.
		numberFormatRule = strings.Replace(numberFormatRule, "$1", ccRule, 1)
	} else {
		// Use the national prefix formatting rule.
		npRule := pattern.NationalPrefixFormattingRule
		if format == FormatNational && npRule != "" {
			// Resolve $NP and $FG in the national prefix formatting rule.
			npRule = resolveFormattingRule(npRule, nationalPrefix)
			// Replace the first $1 in the format with the resolved rule.
			numberFormatRule = strings.Replace(numberFormatRule, "$1", npRule, 1)
		}
	}

	// Apply the pattern regex to format the national number.
	re, err := u.rc.getOrError(pattern.Pattern)
	if err != nil {
		return nationalNumber
	}
	formatted := re.ReplaceAllString(nationalNumber, numberFormatRule)

	if format == FormatRFC3966 {
		// Strip leading punctuation then replace all separators with "-".
		if loc := leadingSeparatorPattern.FindStringIndex(formatted); loc != nil {
			formatted = formatted[loc[1]:]
		}
		formatted = separatorPattern.ReplaceAllString(formatted, "-")
	}

	return formatted
}

// formatNsn formats the national significant number using metadata rules.
func (u *PhoneNumberUtil) formatNsn(number string, meta *metadata.PhoneMetadata, format PhoneNumberFormat, carrierCode string) string {
	// Use intl formats for non-national formatting if available.
	var availableFormats []metadata.NumberFormat
	if len(meta.IntlNumberFormats) == 0 || format == FormatNational {
		availableFormats = meta.NumberFormats
	} else {
		availableFormats = meta.IntlNumberFormats
	}

	pattern := u.chooseFormattingPatternForNumber(availableFormats, number)
	if pattern == nil {
		return number
	}
	return u.formatNsnUsingPatternWithCarrier(number, pattern, format, carrierCode, meta.NationalPrefix)
}

// --- Public formatting methods ---

// Format formats a phone number in the specified format.
func (u *PhoneNumberUtil) Format(number PhoneNumber, format PhoneNumberFormat) string {
	if number.NationalNumber == 0 && number.RawInput != "" {
		return number.RawInput
	}

	nationalSignificantNumber := GetNationalSignificantNumber(number)
	cc := number.CountryCode

	if format == FormatE164 {
		// E164: no NSN formatting needed, just prefix with +CC.
		return prefixNumberWithCountryCallingCode(cc, FormatE164, nationalSignificantNumber)
	}

	if !u.hasValidCountryCallingCode(cc) {
		return nationalSignificantNumber
	}

	regionCode := u.getRegionCodeForCountryCode(cc)
	meta := u.getMetadataForRegionOrCallingCode(cc, regionCode)

	formattedNumber := u.formatNsn(nationalSignificantNumber, meta, format, "")
	formattedNumber = maybeAppendFormattedExtension(number, meta, format, formattedNumber)
	formattedNumber = prefixNumberWithCountryCallingCode(cc, format, formattedNumber)
	return formattedNumber
}

// FormatByPattern formats using caller-supplied formatting patterns.
func (u *PhoneNumberUtil) FormatByPattern(number PhoneNumber, format PhoneNumberFormat, userDefinedFormats []NumberFormatRule) string {
	cc := number.CountryCode
	nationalSignificantNumber := GetNationalSignificantNumber(number)

	if !u.hasValidCountryCallingCode(cc) {
		return nationalSignificantNumber
	}

	regionCode := u.getRegionCodeForCountryCode(cc)
	meta := u.getMetadataForRegionOrCallingCode(cc, regionCode)

	// Convert user formats to internal type with $NP/$FG already resolved.
	internalFormats := make([]metadata.NumberFormat, len(userDefinedFormats))
	for i, uf := range userDefinedFormats {
		internalFormats[i] = metadata.NumberFormat{
			Pattern:               uf.Pattern,
			Format:                uf.Format,
			LeadingDigitsPatterns: uf.LeadingDigitsPatterns,
		}

		npRule := uf.NationalPrefixFormattingRule
		if npRule != "" {
			np := meta.NationalPrefix
			if np != "" {
				// Resolve $NP and $FG for the user-supplied rule.
				npRule = resolveFormattingRule(npRule, np)
				internalFormats[i].NationalPrefixFormattingRule = npRule
			}
			// If no national prefix, leave the rule empty (no prefix to apply).
		}
	}

	pattern := u.chooseFormattingPatternForNumber(internalFormats, nationalSignificantNumber)
	if pattern == nil {
		return nationalSignificantNumber
	}

	formattedNumber := u.formatNsnUsingPatternWithCarrier(nationalSignificantNumber, pattern, format, "", meta.NationalPrefix)
	formattedNumber = maybeAppendFormattedExtension(number, meta, FormatNational, formattedNumber)
	formattedNumber = prefixNumberWithCountryCallingCode(cc, format, formattedNumber)
	return formattedNumber
}

// FormatNationalNumberWithCarrierCode formats in national format with an explicit carrier code.
func (u *PhoneNumberUtil) FormatNationalNumberWithCarrierCode(number PhoneNumber, carrierCode string) string {
	cc := number.CountryCode
	nationalSignificantNumber := GetNationalSignificantNumber(number)

	if !u.hasValidCountryCallingCode(cc) {
		return nationalSignificantNumber
	}

	regionCode := u.getRegionCodeForCountryCode(cc)
	meta := u.getMetadataForRegionOrCallingCode(cc, regionCode)

	formattedNumber := u.formatNsn(nationalSignificantNumber, meta, FormatNational, carrierCode)
	formattedNumber = maybeAppendFormattedExtension(number, meta, FormatNational, formattedNumber)
	formattedNumber = prefixNumberWithCountryCallingCode(cc, FormatNational, formattedNumber)
	return formattedNumber
}

// FormatNationalNumberWithPreferredCarrierCode formats using the number's stored
// carrier code, falling back to the provided code if none is stored.
func (u *PhoneNumberUtil) FormatNationalNumberWithPreferredCarrierCode(number PhoneNumber, fallbackCarrierCode string) string {
	cc := number.PreferredDomesticCarrierCode
	if cc == "" {
		cc = fallbackCarrierCode
	}
	return u.FormatNationalNumberWithCarrierCode(number, cc)
}

// FormatNumberForMobileDialing returns what to dial from a mobile phone in the given region.
// Returns empty string if the number cannot be reached from that region.
func (u *PhoneNumberUtil) FormatNumberForMobileDialing(number PhoneNumber, regionCallingFrom string, withFormatting bool) string {
	cc := number.CountryCode
	if !u.hasValidCountryCallingCode(cc) {
		if number.RawInput != "" {
			return number.RawInput
		}
		return ""
	}

	var formattedNumber string

	// Clear extension for dialing.
	numberNoExt := number
	numberNoExt.Extension = ""

	regionCode := u.getRegionCodeForCountryCode(cc)
	numberType := u.GetNumberType(numberNoExt)
	isValidNumber := numberType != TypeUnknown

	if regionCallingFrom == regionCode {
		isFixedLineOrMobile := numberType == TypeFixedLine ||
			numberType == TypeMobile ||
			numberType == TypeFixedLineOrMobile

		if regionCode == "BR" && isFixedLineOrMobile {
			if numberNoExt.PreferredDomesticCarrierCode != "" {
				formattedNumber = u.FormatNationalNumberWithPreferredCarrierCode(numberNoExt, "")
			}
			// Otherwise formattedNumber stays empty — cannot dial without carrier code.
		} else if cc == nanpaCountryCode {
			// NANPA: use international format if dialable, else national.
			regionMeta := u.getMetadataForRegion(regionCallingFrom)
			nsn := GetNationalSignificantNumber(numberNoExt)
			if u.CanBeInternationallyDialled(numberNoExt) &&
				regionMeta != nil && u.testNumberLength(nsn, regionMeta) != TooShort {
				formattedNumber = u.Format(numberNoExt, FormatInternational)
			} else {
				formattedNumber = u.Format(numberNoExt, FormatNational)
			}
		} else {
			// Non-NANPA same-region cases.
			if (regionCode == regionCodeNonGeo ||
				((regionCode == "MX" || regionCode == "CL" || regionCode == "UZ") && isFixedLineOrMobile)) &&
				u.CanBeInternationallyDialled(numberNoExt) {
				formattedNumber = u.Format(numberNoExt, FormatInternational)
			} else {
				formattedNumber = u.Format(numberNoExt, FormatNational)
			}
		}
	} else if isValidNumber && u.CanBeInternationallyDialled(numberNoExt) {
		if withFormatting {
			formattedNumber = u.Format(numberNoExt, FormatInternational)
		} else {
			formattedNumber = u.Format(numberNoExt, FormatE164)
		}
		return formattedNumber
	}

	if !withFormatting {
		formattedNumber = NormalizeDiallableCharsOnly(formattedNumber)
	}
	return formattedNumber
}

// FormatOutOfCountryCallingNumber formats for dialing from a specific country.
func (u *PhoneNumberUtil) FormatOutOfCountryCallingNumber(number PhoneNumber, callingFrom string) string {
	if !u.isValidRegionCode(callingFrom) {
		return u.Format(number, FormatInternational)
	}

	cc := number.CountryCode
	nationalSignificantNumber := GetNationalSignificantNumber(number)

	if !u.hasValidCountryCallingCode(cc) {
		return nationalSignificantNumber
	}

	if cc == nanpaCountryCode {
		if u.isNANPACountry(callingFrom) {
			// For NANPA regions, prefix national format with country code.
			return strconv.FormatInt(int64(cc), 10) + " " + u.Format(number, FormatNational)
		}
	} else if cc == u.getCountryCodeForValidRegion(callingFrom) {
		// Same country calling code — use national format.
		return u.Format(number, FormatNational)
	}

	metaCallingFrom := u.getMetadataForRegion(callingFrom)
	internationalPrefix := metaCallingFrom.InternationalPrefix

	// Determine the international prefix to use for formatting.
	var intlPrefixForFormatting string
	if metaCallingFrom.PreferredInternationalPrefix != "" {
		intlPrefixForFormatting = metaCallingFrom.PreferredInternationalPrefix
	} else if singleInternationalPrefixPattern.MatchString(internationalPrefix) {
		intlPrefixForFormatting = internationalPrefix
	}

	regionCode := u.getRegionCodeForCountryCode(cc)
	meta := u.getMetadataForRegionOrCallingCode(cc, regionCode)

	formattedNumber := u.formatNsn(nationalSignificantNumber, meta, FormatInternational, "")
	formattedNumber = maybeAppendFormattedExtension(number, meta, FormatInternational, formattedNumber)

	ccStr := strconv.FormatInt(int64(cc), 10)
	if intlPrefixForFormatting != "" {
		return intlPrefixForFormatting + " " + ccStr + " " + formattedNumber
	}
	return prefixNumberWithCountryCallingCode(cc, FormatInternational, formattedNumber)
}

// FormatOutOfCountryKeepingAlphaChars formats for out-of-country dialing
// while preserving alpha characters from the raw input.
func (u *PhoneNumberUtil) FormatOutOfCountryKeepingAlphaChars(number PhoneNumber, callingFrom string) string {
	if number.RawInput == "" {
		return u.FormatOutOfCountryCallingNumber(number, callingFrom)
	}

	cc := number.CountryCode
	if !u.hasValidCountryCallingCode(cc) {
		return number.RawInput
	}

	// Normalize punctuation in raw input, keeping grouping symbols.
	rawInput := normalizeHelper(number.RawInput, true)

	// Trim everything before the first three digits of the national number.
	nationalNumber := GetNationalSignificantNumber(number)
	if len(nationalNumber) > 3 {
		idx := strings.Index(rawInput, nationalNumber[:3])
		if idx >= 0 {
			rawInput = rawInput[idx:]
		}
	}

	metaCallingFrom := u.getMetadataForRegion(callingFrom)

	if cc == nanpaCountryCode {
		if u.isNANPACountry(callingFrom) {
			return strconv.FormatInt(int64(cc), 10) + " " + rawInput
		}
	} else if metaCallingFrom != nil && cc == u.getCountryCodeForValidRegion(callingFrom) {
		// Same country calling code — format using national prefix rules.
		pattern := u.chooseFormattingPatternForNumber(metaCallingFrom.NumberFormats, nationalNumber)
		if pattern == nil {
			return rawInput
		}
		// Create a modified pattern to just concatenate groups after the national prefix.
		modifiedPattern := *pattern
		modifiedPattern.Pattern = `(\d+)(.*)`
		modifiedPattern.Format = "$1$2"
		formatted := u.formatNsnUsingPatternWithCarrier(rawInput, &modifiedPattern, FormatNational, "", metaCallingFrom.NationalPrefix)
		return formatted
	}

	var intlPrefixForFormatting string
	if metaCallingFrom != nil {
		internationalPrefix := metaCallingFrom.InternationalPrefix
		if singleInternationalPrefixPattern.MatchString(internationalPrefix) {
			intlPrefixForFormatting = internationalPrefix
		} else {
			intlPrefixForFormatting = metaCallingFrom.PreferredInternationalPrefix
		}
	}

	ccStr := strconv.FormatInt(int64(cc), 10)
	regionCode := u.getRegionCodeForCountryCode(cc)
	metaForRegion := u.getMetadataForRegionOrCallingCode(cc, regionCode)

	var formattedNumber string
	if intlPrefixForFormatting != "" {
		formattedNumber = intlPrefixForFormatting + " " + ccStr + " " + rawInput
	} else {
		formattedNumber = prefixNumberWithCountryCallingCode(cc, FormatInternational, rawInput)
	}

	// Strip any extension from the formatted number and re-append properly.
	formattedNoExt, _ := u.maybeStripExtension(formattedNumber)
	formattedNumber = maybeAppendFormattedExtension(number, metaForRegion, FormatInternational, formattedNoExt)
	return formattedNumber
}

// FormatInOriginalFormat reconstructs the number in the format it was originally entered.
// Requires ParseAndKeepRawInput to have been used.
func (u *PhoneNumberUtil) FormatInOriginalFormat(number PhoneNumber, regionCallingFrom string) string {
	if number.RawInput != "" && !u.hasFormattingPatternForNumber(number) {
		return number.RawInput
	}

	if number.CountryCodeSource == CountryCodeUnspecified {
		return u.Format(number, FormatNational)
	}

	var formattedNumber string
	switch number.CountryCodeSource {
	case CountryCodeFromNumberWithPlus:
		formattedNumber = u.Format(number, FormatInternational)
	case CountryCodeFromNumberWithIDD:
		formattedNumber = u.FormatOutOfCountryCallingNumber(number, regionCallingFrom)
	case CountryCodeFromNumberWithoutPlus:
		formattedNumber = u.Format(number, FormatInternational)
		// Remove the leading '+'.
		if len(formattedNumber) > 0 && formattedNumber[0] == '+' {
			formattedNumber = formattedNumber[1:]
		}
	case CountryCodeFromDefaultCountry:
		regionCode := u.getRegionCodeForCountryCode(number.CountryCode)
		nationalPrefix := u.getNddPrefixForRegion(regionCode, true)
		if nationalPrefix == "" {
			formattedNumber = u.Format(number, FormatNational)
			break
		}

		if u.rawInputContainsNationalPrefix(number.RawInput, nationalPrefix, regionCode) {
			formattedNumber = u.Format(number, FormatNational)
			break
		}

		// Number was entered without national prefix — format without it.
		meta := u.getMetadataForRegion(regionCode)
		nationalNumber := GetNationalSignificantNumber(number)
		formatRule := u.chooseFormattingPatternForNumber(meta.NumberFormats, nationalNumber)
		if formatRule == nil {
			formattedNumber = u.Format(number, FormatNational)
			break
		}

		candidateRule := formatRule.NationalPrefixFormattingRule
		if candidateRule != "" {
			// Resolve before checking for $1.
			candidateRule = resolveFormattingRule(candidateRule, nationalPrefix)
			idx := strings.Index(candidateRule, "$1")
			if idx < 0 {
				formattedNumber = u.Format(number, FormatNational)
				break
			}
			candidateRule = candidateRule[:idx]
			candidateRule = NormalizeDigitsOnly(candidateRule)
		}
		if candidateRule == "" {
			// National prefix not used for this format.
			formattedNumber = u.Format(number, FormatNational)
			break
		}

		// Format without national prefix by clearing the rule.
		modifiedRule := *formatRule
		modifiedRule.NationalPrefixFormattingRule = ""
		modifiedFormats := []metadata.NumberFormat{modifiedRule}
		pattern := u.chooseFormattingPatternForNumber(modifiedFormats, nationalNumber)
		if pattern == nil {
			formattedNumber = u.Format(number, FormatNational)
		} else {
			formattedNumber = u.formatNsnUsingPatternWithCarrier(nationalNumber, pattern, FormatNational, "", meta.NationalPrefix)
			formattedNumber = maybeAppendFormattedExtension(number, meta, FormatNational, formattedNumber)
			formattedNumber = prefixNumberWithCountryCallingCode(number.CountryCode, FormatNational, formattedNumber)
		}
	default:
		formattedNumber = u.Format(number, FormatNational)
	}

	// Verify formatting didn't change the diallable digits.
	if formattedNumber != "" && number.RawInput != "" {
		normalizedFormatted := NormalizeDiallableCharsOnly(formattedNumber)
		normalizedRaw := NormalizeDiallableCharsOnly(number.RawInput)
		if normalizedFormatted != normalizedRaw {
			return number.RawInput
		}
	}

	return formattedNumber
}

// --- Private helper methods ---

// hasFormattingPatternForNumber returns true if a formatting pattern exists for this number.
func (u *PhoneNumberUtil) hasFormattingPatternForNumber(number PhoneNumber) bool {
	cc := number.CountryCode
	regionCode := u.getRegionCodeForCountryCode(cc)
	meta := u.getMetadataForRegionOrCallingCode(cc, regionCode)
	if meta == nil {
		return false
	}
	nationalNumber := GetNationalSignificantNumber(number)
	return u.chooseFormattingPatternForNumber(meta.NumberFormats, nationalNumber) != nil
}

// isNANPACountry returns true if the region is in the North American Numbering Plan.
func (u *PhoneNumberUtil) isNANPACountry(region string) bool {
	_, ok := metadata.NANPARegions[region]
	return ok
}

// getNddPrefixForRegion returns the national dialing prefix for a region.
func (u *PhoneNumberUtil) getNddPrefixForRegion(region string, stripNonDigits bool) string {
	meta := u.getMetadataForRegion(region)
	if meta == nil {
		return ""
	}
	prefix := meta.NationalPrefix
	if stripNonDigits {
		prefix = strings.ReplaceAll(prefix, "~", "")
	}
	return prefix
}

// rawInputContainsNationalPrefix checks if the raw input starts with the national prefix.
func (u *PhoneNumberUtil) rawInputContainsNationalPrefix(rawInput, nationalPrefix, regionCode string) bool {
	normalizedInput := NormalizeDigitsOnly(rawInput)
	if !strings.HasPrefix(normalizedInput, nationalPrefix) {
		return false
	}
	// Verify that stripping the prefix produces a valid number (avoids false
	// matches like Japanese 00777123 being mistaken for prefix "0" + "0777123").
	withoutPrefix := normalizedInput[len(nationalPrefix):]
	parsed, err := u.Parse(withoutPrefix, regionCode)
	if err != nil {
		return false
	}
	return u.IsValidNumber(parsed)
}

// normalizeHelper normalizes a string for alpha-char formatting.
// When keepNonMappable is true, keeps grouping symbols (spaces, dashes) intact.
func normalizeHelper(number string, keepNonMappable bool) string {
	var sb strings.Builder
	sb.Grow(len(number))
	for _, ch := range number {
		if d, ok := alphaToDigit[ch]; ok {
			sb.WriteRune(d)
		} else if ch >= '0' && ch <= '9' {
			sb.WriteRune(ch)
		} else if keepNonMappable {
			// Keep grouping symbols like spaces, dashes, parens, dots, slashes.
			switch {
			case ch == ' ' || ch == '(' || ch == ')' || ch == '.' || ch == '/' || ch == '+':
				sb.WriteRune(ch)
			case ch == '-' || (ch >= 0x2010 && ch <= 0x2015) || ch == 0x2212 ||
				ch == 0x30FC || (ch >= 0xFF0D && ch <= 0xFF0F):
				sb.WriteRune(ch)
			case ch == 0x00A0 || ch == 0x3000: // non-breaking space, ideographic space
				sb.WriteRune(ch)
			}
		}
	}
	return sb.String()
}

// --- Public helpers that the spec requires ---

// IsNANPACountry returns true if the region is in the North American Numbering Plan.
func (u *PhoneNumberUtil) IsNANPACountry(region string) bool {
	return u.isNANPACountry(region)
}

// GetNddPrefixForRegion returns the national dialing prefix for a region.
func (u *PhoneNumberUtil) GetNddPrefixForRegion(region string, stripNonDigits bool) string {
	return u.getNddPrefixForRegion(region, stripNonDigits)
}

// GetCountryMobileToken returns the mobile token for countries that require one
// when dialing internationally (e.g., Argentina requires "9").
func GetCountryMobileToken(cc int32) string {
	if token, ok := mobileTokenMap[cc]; ok {
		return token
	}
	return ""
}

// mobileTokenMap maps country calling codes to their mobile tokens.
var mobileTokenMap = map[int32]string{
	54: "9", // Argentina
}

// GetExampleNumber returns a valid example number for the given region.
// Returns false if no example is available.
func (u *PhoneNumberUtil) GetExampleNumber(region string) (PhoneNumber, bool) {
	return u.GetExampleNumberForType(region, TypeFixedLine)
}

// GetExampleNumberForType returns a valid example number of the given type
// for the region. Falls back to other types if the requested type has no example.
// Returns false if no example is available.
func (u *PhoneNumberUtil) GetExampleNumberForType(region string, numType PhoneNumberType) (PhoneNumber, bool) {
	var meta *metadata.PhoneMetadata
	if region == regionCodeNonGeo {
		return PhoneNumber{}, false
	}
	meta = u.getMetadataForRegion(region)
	if meta == nil {
		return PhoneNumber{}, false
	}

	desc := getDescForType(meta, numType)
	if desc.ExampleNumber != "" {
		num, err := u.Parse(desc.ExampleNumber, region)
		if err == nil {
			return num, true
		}
	}

	// Fallback: try other types.
	fallbacks := []PhoneNumberType{
		TypeFixedLine, TypeMobile, TypeTollFree, TypePremiumRate,
		TypeSharedCost, TypeVOIP, TypePersonalNumber, TypePager, TypeUAN, TypeVoicemail,
	}
	for _, ft := range fallbacks {
		if ft == numType {
			continue
		}
		d := getDescForType(meta, ft)
		if d.ExampleNumber != "" {
			num, err := u.Parse(d.ExampleNumber, region)
			if err == nil {
				return num, true
			}
		}
	}
	return PhoneNumber{}, false
}

// GetInvalidExampleNumber returns an example of an invalid number for the region.
// Useful for testing validation logic. Returns false if unable to generate one.
func (u *PhoneNumberUtil) GetInvalidExampleNumber(region string) (PhoneNumber, bool) {
	// Start with the example number and modify it to be invalid.
	num, ok := u.GetExampleNumber(region)
	if !ok {
		return PhoneNumber{}, false
	}

	nsn := GetNationalSignificantNumber(num)
	// Try appending digits until we get an invalid number.
	for i := 0; i < 10; i++ {
		nsn += "1"
		candidate, err := u.Parse("+"+strconv.FormatInt(int64(num.CountryCode), 10)+nsn, "ZZ")
		if err != nil {
			continue
		}
		if !u.IsValidNumber(candidate) {
			return candidate, true
		}
	}
	return PhoneNumber{}, false
}

// TruncateTooLongNumber attempts to shorten a too-long number to a valid length
// by removing trailing digits. Returns the truncated number and true on success,
// or the original number and false if truncation cannot produce a valid number.
func (u *PhoneNumberUtil) TruncateTooLongNumber(number PhoneNumber) (PhoneNumber, bool) {
	if u.IsValidNumber(number) {
		return number, true
	}

	candidate := number
	nsn := GetNationalSignificantNumber(number)
	for len(nsn) > minLengthForNSN {
		// Remove last digit.
		nsn = nsn[:len(nsn)-1]
		candidate.NationalNumber = 0
		candidate.ItalianLeadingZero = false
		candidate.NumberOfLeadingZeros = 0

		// Reconstruct from the trimmed NSN.
		startIdx := 0
		if len(nsn) > 0 && nsn[0] == '0' {
			candidate.ItalianLeadingZero = true
			zeros := 0
			for startIdx < len(nsn) && nsn[startIdx] == '0' {
				zeros++
				startIdx++
			}
			candidate.NumberOfLeadingZeros = int32(zeros)
		}

		if startIdx < len(nsn) {
			nn, err := strconv.ParseUint(nsn[startIdx:], 10, 64)
			if err != nil {
				return number, false
			}
			candidate.NationalNumber = nn
		}

		if u.IsPossibleNumber(candidate) && u.IsValidNumber(candidate) {
			return candidate, true
		}
	}
	return number, false
}
