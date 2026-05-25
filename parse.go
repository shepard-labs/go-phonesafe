package phonesafe

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/shepard-labs/go-phonesafe/internal/metadata"
)

// Constants matching upstream libphonenumber.
const (
	minLengthForNSN  = 2
	maxLengthForNSN  = 17
	maxLengthCC      = 3
	nanpaCountryCode = 1
	regionCodeNonGeo = "001"
)

// RFC3966 constants.
const (
	rfc3966Prefix         = "tel:"
	rfc3966PhoneContext   = ";phone-context="
	rfc3966ISDNSubaddress = ";isub="
	rfc3966ExtnPrefix     = ";ext="
)

// Pre-compiled patterns used across parse functions.
var (
	// validStartCharPattern matches the first valid character in a phone number.
	validStartCharPattern *regexp.Regexp

	// unwantedEndCharPattern matches trailing characters to strip.
	unwantedEndCharPattern *regexp.Regexp

	// secondNumberStartPattern matches the start of a second number (for extraction).
	secondNumberStartPattern *regexp.Regexp

	// validPhoneNumberPattern is a broad pattern for viable phone numbers.
	validPhoneNumberPattern *regexp.Regexp

	// plusCharsPattern matches leading plus characters.
	plusCharsPattern *regexp.Regexp

	// extnPattern matches extension markers at end of number (for parsing).
	extnPattern *regexp.Regexp

	// validAlphaPattern matches numbers with 3+ alpha chars (for Normalize).
	validAlphaPattern *regexp.Regexp

	// capturingDigitPattern matches a single digit (used by parsePrefixAsIDD).
	capturingDigitPattern *regexp.Regexp

	// rfc3966GlobalNumberDigits validates global phone-context value.
	rfc3966GlobalNumberDigitsPattern *regexp.Regexp

	// rfc3966DomainNamePattern validates domain phone-context value.
	rfc3966DomainNamePattern *regexp.Regexp
)

func init() {
	// Plus chars: ASCII + and fullwidth ＋
	const plusChars = `+\x{FF0B}`

	// Valid punctuation found in phone numbers.
	// Dash is at start of class to be literal. 'x' is literal.
	// Ranges: U+2010-U+2015, U+FF0D-U+FF0F.
	const validPunct = `\x{2010}-\x{2015}\x{2212}\x{30FC}\x{FF0D}-\x{FF0F}` +
		` \x{00A0}\x{00AD}\x{200B}\x{2060}\x{3000}` +
		`()\x{FF08}\x{FF09}\x{FF3B}\x{FF3D}\x{2053}\x{223C}` +
		`x.\\\/~\-`

	const digits = `\p{Nd}`
	const validAlpha = `a-zA-Z`

	validStartCharPattern = regexp.MustCompile(`[` + plusChars + digits + `]`)
	unwantedEndCharPattern = regexp.MustCompile(`[^\p{N}\p{L}#]+$`)
	secondNumberStartPattern = regexp.MustCompile(`[\\\\\/] *x`)
	plusCharsPattern = regexp.MustCompile(`^[` + plusChars + `]+`)
	capturingDigitPattern = regexp.MustCompile(`(` + `[` + digits + `]` + `)`)

	// Extension patterns: 6 capture groups for different extension formats.
	// Build this first since validPhoneNumberPattern needs the extension pattern.
	extnPatternStr := buildExtnPatternStr()
	extnPattern = regexp.MustCompile(extnPatternStr + `$`)

	// A valid phone number has at minimum 2 digits, or starts with valid punctuation
	// sequences interspersed with at least 3 groups of digits.
	// Includes optional extension suffix (upstream valid_phone_number_pattern_).
	// Anchored for full match (upstream uses FullMatch).
	validPhoneNumberPattern = regexp.MustCompile(
		`^(?i)(?:[` + digits + `]{2}|` +
			`[` + plusChars + `]*(?:[` + validPunct + `]*[` + digits + `]){3,}` +
			`[` + validPunct + `*` + validAlpha + digits + `]*)` +
			`(?:` + extnPatternStr + `)?$`)

	// Numbers with 3+ alpha chars may be vanity numbers.
	validAlphaPattern = regexp.MustCompile(`(?i)(?:.*?[A-Za-z]){3}`)

	// RFC3966 phone-context validation patterns.
	rfc3966GlobalNumberDigitsPattern = regexp.MustCompile(
		`^\+[` + digits + `\-.\x{FF0D}\x{FF0E}\x{FF08}\x{FF09}` + `]+$`)
	rfc3966DomainNamePattern = regexp.MustCompile(
		`^([a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
}

// buildExtnPatternStr constructs the extension matching regex string for parsing.
func buildExtnPatternStr() string {
	const digits = `\p{Nd}`

	extnDigits := func(maxLen int) string {
		return `([` + digits + `]{1,` + strconv.Itoa(maxLen) + `})`
	}

	// Separator between number and extension label.
	separatorBetween := `[ \x{00A0}\t,]*`
	// Possible chars after extension label.
	charsAfterLabel := `[:\.\x{FF0E}]?[ \x{00A0}\t,\-]*`
	optionalSuffix := `#?`

	// Explicit labels: ext, extn, extension, anexo (and full-width variants).
	explicitLabels := `(?:e?xt(?:ensi(?:o\x{0301}?|\x{00F3}))?n?|` +
		`(?:\x{FF45})?\x{FF58}\x{FF54}(?:\x{FF4E})?|` +
		`\x{0434}\x{043E}\x{0431}|anexo)`
	// Ambiguous labels: x, #, ~, int
	ambiguousLabels := `(?:[x\x{FF58}#\x{FF03}~\x{FF5E}]|int|\x{FF49}\x{FF4E}\x{FF54})`

	// Group 1: RFC3966 ;ext=
	rfc := rfc3966ExtnPrefix + extnDigits(20)
	// Group 2: Explicit labels
	explicit := separatorBetween + explicitLabels + charsAfterLabel + extnDigits(20) + optionalSuffix
	// Group 3: Ambiguous labels
	ambiguous := separatorBetween + ambiguousLabels + charsAfterLabel + extnDigits(9) + optionalSuffix
	// Group 4: American style: "- 503#"
	american := `[- ]+` + extnDigits(6) + `#`
	// Group 5: Auto-dialling (,, or ;)
	autoDialSep := `[ \x{00A0}\t]*`
	autoDial := autoDialSep + `(?:,,|;)` + charsAfterLabel + extnDigits(15) + optionalSuffix
	// Group 6: Only commas
	onlyCommas := autoDialSep + `(?:,)+` + charsAfterLabel + extnDigits(9) + optionalSuffix

	pattern := `(?i)(?:` + rfc + `|` + explicit + `|` + ambiguous + `|` + american + `|` + autoDial + `|` + onlyCommas + `)`
	return pattern
}

// --- Metadata lookup helpers ---

func (u *PhoneNumberUtil) getMetadataForRegion(regionCode string) *metadata.PhoneMetadata {
	return metadata.RegionMetadata[regionCode]
}

func (u *PhoneNumberUtil) getMetadataForNonGeoRegion(cc int32) *metadata.PhoneMetadata {
	return metadata.NonGeoMetadata[cc]
}

func (u *PhoneNumberUtil) getMetadataForRegionOrCallingCode(cc int32, region string) *metadata.PhoneMetadata {
	if region == regionCodeNonGeo {
		return u.getMetadataForNonGeoRegion(cc)
	}
	return u.getMetadataForRegion(region)
}

func (u *PhoneNumberUtil) getRegionCodeForCountryCode(cc int32) string {
	regions := metadata.CountryCodeToRegions[cc]
	if len(regions) == 0 {
		// Check if this is a non-geographical entity.
		if _, ok := metadata.NonGeoMetadata[cc]; ok {
			return regionCodeNonGeo
		}
		return "ZZ"
	}
	return regions[0]
}

func (u *PhoneNumberUtil) hasValidCountryCallingCode(cc int32) bool {
	if _, ok := metadata.CountryCodeToRegions[cc]; ok {
		return true
	}
	_, ok := metadata.NonGeoMetadata[cc]
	return ok
}

func (u *PhoneNumberUtil) isValidRegionCode(region string) bool {
	if region == "" || region == "ZZ" || region == regionCodeNonGeo {
		return false
	}
	_, ok := metadata.RegionMetadata[region]
	return ok
}

// GetNationalSignificantNumber returns the national significant number string
// for a parsed PhoneNumber, including Italian leading zeros.
func GetNationalSignificantNumber(number PhoneNumber) string {
	nsn := strconv.FormatUint(number.NationalNumber, 10)
	if number.ItalianLeadingZero {
		zeros := int(number.NumberOfLeadingZeros)
		if zeros == 0 {
			zeros = 1
		}
		nsn = strings.Repeat("0", zeros) + nsn
	}
	return nsn
}

// --- Parse pipeline functions ---

// extractPossibleNumber strips leading non-phone characters and trailing junk.
func (u *PhoneNumberUtil) extractPossibleNumber(number string) string {
	if !utf8.ValidString(number) {
		return ""
	}

	// Find first valid start character.
	loc := validStartCharPattern.FindStringIndex(number)
	if loc == nil {
		return ""
	}
	number = number[loc[0]:]

	// Trim unwanted end characters.
	number = unwantedEndCharPattern.ReplaceAllString(number, "")
	if number == "" {
		return ""
	}

	// Remove second number if present (e.g., "6985 x302/x2303" → "6985 x302").
	if loc := secondNumberStartPattern.FindStringIndex(number); loc != nil {
		number = number[:loc[0]]
	}

	return number
}

// isViablePhoneNumber checks whether a string could possibly be a phone number.
func (u *PhoneNumberUtil) isViablePhoneNumber(number string) bool {
	if len(number) < minLengthForNSN {
		return false
	}
	return validPhoneNumberPattern.MatchString(number)
}

// checkRegionForParsing validates that either the region is valid or the
// number starts with a plus (indicating international format).
func (u *PhoneNumberUtil) checkRegionForParsing(number, defaultRegion string) bool {
	if u.isValidRegionCode(defaultRegion) {
		return true
	}
	if number == "" {
		return false
	}
	return plusCharsPattern.MatchString(number)
}

// maybeStripExtension attempts to strip an extension from the end of the number.
// Returns the number (possibly stripped) and the extension (empty if none found).
func (u *PhoneNumberUtil) maybeStripExtension(number string) (string, string) {
	matches := extnPattern.FindStringSubmatchIndex(number)
	if matches == nil {
		return number, ""
	}

	// The match starts at matches[0] — everything before it is the phone number.
	stripped := number[:matches[0]]

	// Check that the remainder is still a viable phone number.
	if !u.isViablePhoneNumber(stripped) {
		return number, ""
	}

	// Find the first non-empty capture group (groups are at indices 2,3 / 4,5 / ...).
	for i := 2; i < len(matches); i += 2 {
		if matches[i] != -1 {
			return stripped, number[matches[i]:matches[i+1]]
		}
	}
	return number, ""
}

// maybeStripInternationalPrefixAndNormalize strips + or IDD prefix, normalizes.
func (u *PhoneNumberUtil) maybeStripInternationalPrefixAndNormalize(
	possibleIDDPrefix string, number string) (string, CountryCodeSource) {

	if number == "" {
		return number, CountryCodeFromDefaultCountry
	}

	// Check for leading plus sign.
	if loc := plusCharsPattern.FindStringIndex(number); loc != nil {
		number = number[loc[1]:]
		number = normalize(number)
		return number, CountryCodeFromNumberWithPlus
	}

	// Normalize before trying IDD.
	number = normalize(number)

	// Try to match the IDD pattern at the start.
	if possibleIDDPrefix == "" || possibleIDDPrefix == "NonMatch" {
		return number, CountryCodeFromDefaultCountry
	}

	if result, ok := u.parsePrefixAsIDD(possibleIDDPrefix, number); ok {
		return result, CountryCodeFromNumberWithIDD
	}
	return number, CountryCodeFromDefaultCountry
}

// parsePrefixAsIDD strips the IDD prefix if present and valid.
func (u *PhoneNumberUtil) parsePrefixAsIDD(iddPatternStr string, number string) (string, bool) {
	iddPattern, err := u.rc.getOrError(`^(?:` + iddPatternStr + `)`)
	if err != nil {
		return number, false
	}

	loc := iddPattern.FindStringIndex(number)
	if loc == nil {
		return number, false
	}

	// After stripping IDD, first digit must NOT be '0' (country codes can't start with 0).
	afterIDD := number[loc[1]:]
	if digitLoc := capturingDigitPattern.FindString(afterIDD); digitLoc != "" {
		normalized := NormalizeDigitsOnly(digitLoc)
		if normalized == "0" {
			return number, false
		}
	}

	return afterIDD, true
}

// extractCountryCode extracts a 1-3 digit country calling code from the
// beginning of the number. Returns the code and remaining number. Returns 0
// if no valid country code is found.
func (u *PhoneNumberUtil) extractCountryCode(nationalNumber string) (int32, string) {
	if nationalNumber == "" || nationalNumber[0] == '0' {
		return 0, nationalNumber
	}

	maxLen := min(len(nationalNumber), maxLengthCC)

	for i := 1; i <= maxLen; i++ {
		code, err := strconv.Atoi(nationalNumber[:i])
		if err != nil {
			continue
		}
		cc := int32(code)
		if u.hasValidCountryCallingCode(cc) {
			return cc, nationalNumber[i:]
		}
	}
	return 0, nationalNumber
}

// maybeExtractCountryCode tries to extract a country calling code.
// It modifies the national number to remove any international prefix/code.
func (u *PhoneNumberUtil) maybeExtractCountryCode(
	defaultMeta *metadata.PhoneMetadata,
	keepRawInput bool,
	nationalNumber string,
) (normalizedNumber string, cc int32, ccs CountryCodeSource, err error) {

	possibleIDD := "NonMatch"
	if defaultMeta != nil {
		possibleIDD = defaultMeta.InternationalPrefix
	}

	ccs = CountryCodeFromDefaultCountry
	normalizedNumber, ccs = u.maybeStripInternationalPrefixAndNormalize(possibleIDD, nationalNumber)

	if ccs != CountryCodeFromDefaultCountry {
		// Number had international prefix — extract CC.
		if len(normalizedNumber) <= minLengthForNSN {
			return "", 0, ccs, &ParseError{
				Code:    ErrTooShortAfterIDD,
				Message: "phone number is too short after IDD",
			}
		}

		cc, remainder := u.extractCountryCode(normalizedNumber)
		if cc != 0 {
			return remainder, cc, ccs, nil
		}
		return "", 0, ccs, &ParseError{
			Code:    ErrInvalidCountryCode,
			Message: "country calling code not recognized",
		}
	}

	// No international prefix — try using default region's country code.
	if defaultMeta != nil {
		defaultCC := defaultMeta.CountryCode
		defaultCCStr := strconv.FormatInt(int64(defaultCC), 10)

		if strings.HasPrefix(normalizedNumber, defaultCCStr) {
			potentialNational := normalizedNumber[len(defaultCCStr):]

			// Try stripping national prefix from potential number.
			strippedPotential, _, _ := u.maybeStripNationalPrefixAndCarrierCode(defaultMeta, potentialNational)

			// Check if the number is better with the CC stripped.
			generalPattern := defaultMeta.GeneralDesc.NationalNumberPattern
			if generalPattern != "" {
				origMatches := u.matchesEntirely(generalPattern, normalizedNumber)
				potentialMatches := u.matchesEntirely(generalPattern, strippedPotential)

				if (!origMatches && potentialMatches) ||
					u.testNumberLength(normalizedNumber, defaultMeta) == TooLong {
					normalizedNumber = potentialNational
					cc = defaultCC
					if keepRawInput {
						ccs = CountryCodeFromNumberWithoutPlus
					}
					return normalizedNumber, cc, ccs, nil
				}
			}
		}
	}

	// No country code found.
	return normalizedNumber, 0, ccs, nil
}

// maybeStripNationalPrefixAndCarrierCode strips the national prefix and
// extracts carrier code if present.
func (u *PhoneNumberUtil) maybeStripNationalPrefixAndCarrierCode(
	meta *metadata.PhoneMetadata, number string) (string, string, bool) {

	npForParsing := meta.NationalPrefixForParsing
	if number == "" || npForParsing == "" {
		return number, "", false
	}

	// Compile the national prefix pattern (anchored at start).
	prefixPattern, err := u.rc.getOrError(`^(?:` + npForParsing + `)`)
	if err != nil {
		return number, "", false
	}

	matches := prefixPattern.FindStringSubmatch(number)
	if matches == nil {
		return number, "", false
	}

	// Check if the entire match is the whole number (nothing would remain).
	matchEnd := len(matches[0])
	if matchEnd == 0 {
		return number, "", false
	}

	transformRule := meta.NationalPrefixTransformRule
	generalPattern := meta.GeneralDesc.NationalNumberPattern

	// Check if original number matches the general pattern.
	isViableOriginal := generalPattern != "" && u.matchesEntirely(generalPattern, number)

	var transformed string
	var carrierCode string

	if transformRule != "" && len(matches) > 1 && matches[1] != "" {
		// Apply transform rule using captured groups.
		transformed = prefixPattern.ReplaceAllString(number, transformRule)
		// Extract carrier code from second group if present.
		if len(matches) > 2 && matches[2] != "" {
			carrierCode = matches[2]
		} else if len(matches) > 1 {
			carrierCode = matches[1]
		}
	} else {
		// Simple strip: remove the matched prefix.
		transformed = number[matchEnd:]
		// Carrier code is the first capture group (if any).
		if len(matches) > 1 {
			carrierCode = matches[1]
		}
	}

	// Validate: if the original matched but the transformed doesn't, reject.
	if isViableOriginal && generalPattern != "" && !u.matchesEntirely(generalPattern, transformed) {
		return number, "", false
	}

	return transformed, carrierCode, true
}

// setItalianLeadingZeros detects and sets leading zeros on a PhoneNumber.
func setItalianLeadingZeros(nationalNumber string, pn *PhoneNumber) {
	if len(nationalNumber) <= 1 || nationalNumber[0] != '0' {
		return
	}

	pn.ItalianLeadingZero = true
	numZeros := int32(1)
	// Count consecutive zeros, but not the last digit if all zeros.
	for i := 1; i < len(nationalNumber)-1; i++ {
		if nationalNumber[i] != '0' {
			break
		}
		numZeros++
	}
	if numZeros != 1 {
		pn.NumberOfLeadingZeros = numZeros
	}
}

// testNumberLength checks the length of a number against metadata possible lengths.
func (u *PhoneNumberUtil) testNumberLength(number string, meta *metadata.PhoneMetadata) ValidationResult {
	possibleLengths := meta.GeneralDesc.PossibleLengths
	if len(possibleLengths) == 0 {
		// No length info — accept anything in NSN range.
		l := int32(len(number))
		if l < minLengthForNSN {
			return TooShort
		}
		if l > maxLengthForNSN {
			return TooLong
		}
		return IsPossible
	}

	localLengths := meta.GeneralDesc.PossibleLengthsLocal

	// -1 indicates the type is not supported.
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

// matchesEntirely checks if the number fully matches the pattern.
func (u *PhoneNumberUtil) matchesEntirely(pattern, number string) bool {
	re, err := u.rc.getOrError(`^(?:` + pattern + `)$`)
	if err != nil {
		return false
	}
	return re.MatchString(number)
}

// --- BuildNationalNumberForParsing + RFC3966 ---

// buildNationalNumberForParsing handles RFC3966 and general number extraction.
func (u *PhoneNumberUtil) buildNationalNumberForParsing(numberToParse string) (string, error) {
	phoneContextIdx := strings.Index(numberToParse, rfc3966PhoneContext)

	if phoneContextIdx >= 0 {
		// Extract phone-context value.
		phoneContext := u.extractPhoneContext(numberToParse, phoneContextIdx)
		if !u.isPhoneContextValid(phoneContext) {
			return "", &ParseError{
				Code:    ErrNotANumber,
				Message: "invalid phone-context",
			}
		}

		var nationalNumber strings.Builder
		if phoneContext != "" && phoneContext[0] == '+' {
			nationalNumber.WriteString(phoneContext)
		}

		// Append portion between "tel:" and ";phone-context=".
		prefixIdx := strings.Index(numberToParse, rfc3966Prefix)
		startIdx := 0
		if prefixIdx >= 0 {
			startIdx = prefixIdx + len(rfc3966Prefix)
		}
		nationalNumber.WriteString(numberToParse[startIdx:phoneContextIdx])

		result := nationalNumber.String()

		// Strip ISDN subaddress.
		if isdnIdx := strings.Index(result, rfc3966ISDNSubaddress); isdnIdx >= 0 {
			result = result[:isdnIdx]
		}
		return result, nil
	}

	// Not RFC3966 — extract possible number.
	result := u.extractPossibleNumber(numberToParse)
	// Strip ISDN subaddress if present.
	if isdnIdx := strings.Index(result, rfc3966ISDNSubaddress); isdnIdx >= 0 {
		result = result[:isdnIdx]
	}
	return result, nil
}

// extractPhoneContext extracts the phone-context parameter value.
func (u *PhoneNumberUtil) extractPhoneContext(number string, idx int) string {
	start := idx + len(rfc3966PhoneContext)
	if start >= len(number) {
		return ""
	}
	// phone-context ends at next ';' or end of string.
	end := strings.IndexByte(number[start:], ';')
	if end >= 0 {
		return number[start : start+end]
	}
	return number[start:]
}

// isPhoneContextValid validates a phone-context value.
func (u *PhoneNumberUtil) isPhoneContextValid(phoneContext string) bool {
	if phoneContext == "" {
		return false
	}
	return rfc3966GlobalNumberDigitsPattern.MatchString(phoneContext) ||
		rfc3966DomainNamePattern.MatchString(phoneContext)
}

// --- Main parse helper ---

// parseHelper is the core parsing implementation.
func (u *PhoneNumberUtil) parseHelper(
	numberToParse, defaultRegion string,
	keepRawInput, checkRegion bool,
) (PhoneNumber, error) {
	var pn PhoneNumber

	// Step 1: Build national number for parsing (handles RFC3966).
	nationalNumber, err := u.buildNationalNumberForParsing(numberToParse)
	if err != nil {
		return pn, err
	}

	// Step 2: Viability check.
	if !u.isViablePhoneNumber(nationalNumber) {
		return pn, &ParseError{
			Code:    ErrNotANumber,
			Message: "the string supplied did not seem to be a phone number",
		}
	}

	// Step 3: Region validation.
	if checkRegion && !u.checkRegionForParsing(nationalNumber, defaultRegion) {
		return pn, &ParseError{
			Code:    ErrInvalidCountryCode,
			Message: "missing or invalid default region",
		}
	}

	if keepRawInput {
		pn.RawInput = numberToParse
	}

	// Step 4: Strip extension (before country code extraction).
	nationalNumber, ext := u.maybeStripExtension(nationalNumber)
	if ext != "" {
		pn.Extension = ext
	}

	// Step 5: Extract country code.
	countryMeta := u.getMetadataForRegion(defaultRegion)
	normalizedNumber, cc, ccs, extractErr := u.maybeExtractCountryCode(
		countryMeta, keepRawInput, nationalNumber)

	if extractErr != nil {
		// If INVALID_COUNTRY_CODE and number starts with +, strip + and retry.
		pe, ok := extractErr.(*ParseError)
		if ok && pe.Code == ErrInvalidCountryCode && plusCharsPattern.MatchString(nationalNumber) {
			withoutPlus := plusCharsPattern.ReplaceAllString(nationalNumber, "")
			normalizedNumber, cc, ccs, extractErr = u.maybeExtractCountryCode(
				countryMeta, keepRawInput, withoutPlus)
			if cc == 0 {
				return pn, &ParseError{
					Code:    ErrInvalidCountryCode,
					Message: "country calling code not recognized",
				}
			}
		} else {
			return pn, extractErr
		}
	}

	if keepRawInput {
		pn.CountryCodeSource = ccs
	}

	// Step 6: Re-lookup metadata if country code changed region.
	var phoneNumberRegion string
	if cc != 0 {
		phoneNumberRegion = u.getRegionCodeForCountryCode(cc)
		if phoneNumberRegion != defaultRegion {
			countryMeta = u.getMetadataForRegionOrCallingCode(cc, phoneNumberRegion)
		}
	} else if countryMeta != nil {
		cc = countryMeta.CountryCode
	}

	// Step 7: Check min length.
	if len(normalizedNumber) < minLengthForNSN {
		return pn, &ParseError{
			Code:    ErrTooShortNSN,
			Message: "the string supplied is too short to be a phone number",
		}
	}

	// Step 8: Strip national prefix and carrier code.
	if countryMeta != nil {
		potentialNational, carrierCode, didStrip := u.maybeStripNationalPrefixAndCarrierCode(
			countryMeta, normalizedNumber)

		// Only accept strip if number length validates.
		if didStrip {
			vr := u.testNumberLength(potentialNational, countryMeta)
			if vr != TooShort && vr != IsPossibleLocalOnly && vr != InvalidLength {
				normalizedNumber = potentialNational
				if keepRawInput && carrierCode != "" {
					pn.PreferredDomesticCarrierCode = carrierCode
				}
			}
		}
	}

	// Step 9: Final length checks.
	if len(normalizedNumber) < minLengthForNSN {
		return pn, &ParseError{
			Code:    ErrTooShortNSN,
			Message: "the string supplied is too short to be a phone number",
		}
	}
	if len(normalizedNumber) > maxLengthForNSN {
		return pn, &ParseError{
			Code:    ErrTooLongNSN,
			Message: "the string supplied is too long to be a phone number",
		}
	}

	// Step 10: Set country code and Italian leading zeros.
	pn.CountryCode = cc
	setItalianLeadingZeros(normalizedNumber, &pn)

	// Step 11: Convert to uint64.
	natNum, _ := strconv.ParseUint(normalizedNumber, 10, 64)
	pn.NationalNumber = natNum

	return pn, nil
}
