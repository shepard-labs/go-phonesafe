package phonesafe

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/shepard-labs/go-phonesafe/internal/metadata"
)

// AsYouTypeFormatter constants.
const (
	separatorBeforeNationalNumber = ' '
	minLeadingDigitsLength        = 3
	digitPlaceholder              = ' ' // punctuation space
)

// Package-level patterns for AsYouTypeFormatter.
var (
	// eligibleFormatPattern checks that a format string contains $1 (and optionally
	// more $N groups) separated by valid phone punctuation. This prevents invalid
	// chars (like * in Israeli star numbers) from leaking into formatted output.
	eligibleFormatPattern *regexp.Regexp

	// nationalPrefixSeparatorsPattern detects space or dash in formatting rules,
	// indicating we should separate the national prefix from the number.
	nationalPrefixSeparatorsPattern *regexp.Regexp

	// firstGroupOnlyPrefixPattern matches NP formatting rules that only reference $1,
	// meaning the rule is essentially a passthrough.
	firstGroupOnlyPrefixPattern *regexp.Regexp

	// digitPlaceholderPattern matches the punctuation space placeholder.
	digitPlaceholderPattern *regexp.Regexp

	// emptyMetadata is a sentinel for invalid regions.
	emptyMetadata = &metadata.PhoneMetadata{InternationalPrefix: "NA"}
)

func init() {
	// Valid punctuation character class (same as parse.go).
	const validPunct = `\-\x{2010}-\x{2015}\x{2212}\x{30FC}\x{FF0D}-\x{FF0F}` +
		` \x{00A0}\x{00AD}\x{200B}\x{2060}\x{3000}` +
		`()\x{FF08}\x{FF09}\x{FF3B}\x{FF3D}\x{2053}\x{223C}` +
		`x.\\\/~`

	eligibleFormatPattern = regexp.MustCompile(
		`^[` + validPunct + `]*\$1[` + validPunct + `]*(\$\d[` + validPunct + `]*)*$`)

	nationalPrefixSeparatorsPattern = regexp.MustCompile(`[- ]`)
	firstGroupOnlyPrefixPattern = regexp.MustCompile(`^\(?\$1\)?$`)
	digitPlaceholderPattern = regexp.MustCompile(string(digitPlaceholder))
}

// AsYouTypeFormatter formats phone numbers incrementally as digits are entered.
// NOT goroutine-safe — use one instance per input context.
type AsYouTypeFormatter struct {
	util           *PhoneNumberUtil
	defaultCountry string

	defaultMetadata *metadata.PhoneMetadata
	currentMetadata *metadata.PhoneMetadata

	currentOutput            string
	formattingTemplate       []rune
	currentFormattingPattern string

	accruedInput                  strings.Builder
	accruedInputWithoutFormatting strings.Builder
	nationalNumber                strings.Builder
	prefixBeforeNationalNumber    strings.Builder

	extractedNationalPrefix string
	possibleFormats         []*metadata.NumberFormat

	ableToFormat                    bool
	inputHasFormatting              bool
	isCompleteNumber                bool
	isExpectingCountryCallingCode   bool
	shouldAddSpaceAfterNationalPrefix bool

	lastMatchPosition  int
	originalPosition   int
	positionToRemember int
}

// NewAsYouTypeFormatter creates an AsYouTypeFormatter for the given region.
func (u *PhoneNumberUtil) NewAsYouTypeFormatter(regionCode string) *AsYouTypeFormatter {
	f := &AsYouTypeFormatter{
		util:           u,
		defaultCountry: regionCode,
		ableToFormat:   true,
	}
	f.currentMetadata = f.aytfGetMetadataForRegion(regionCode)
	f.defaultMetadata = f.currentMetadata
	return f
}

// aytfGetMetadataForRegion returns metadata for the main region of a country code.
func (f *AsYouTypeFormatter) aytfGetMetadataForRegion(regionCode string) *metadata.PhoneMetadata {
	cc := f.util.GetCountryCodeForRegion(regionCode)
	mainCountry := f.util.getRegionCodeForCountryCode(cc)
	meta := f.util.getMetadataForRegion(mainCountry)
	if meta != nil {
		return meta
	}
	return emptyMetadata
}

// Clear resets the formatter for a new number.
func (f *AsYouTypeFormatter) Clear() {
	f.currentOutput = ""
	f.accruedInput.Reset()
	f.accruedInputWithoutFormatting.Reset()
	f.formattingTemplate = f.formattingTemplate[:0]
	f.lastMatchPosition = 0
	f.currentFormattingPattern = ""
	f.prefixBeforeNationalNumber.Reset()
	f.extractedNationalPrefix = ""
	f.nationalNumber.Reset()
	f.ableToFormat = true
	f.inputHasFormatting = false
	f.positionToRemember = 0
	f.originalPosition = 0
	f.isCompleteNumber = false
	f.isExpectingCountryCallingCode = false
	f.possibleFormats = f.possibleFormats[:0]
	f.shouldAddSpaceAfterNationalPrefix = false
	if f.currentMetadata != f.defaultMetadata {
		f.currentMetadata = f.aytfGetMetadataForRegion(f.defaultCountry)
	}
}

// InputDigit feeds the next character and returns the current formatted output.
func (f *AsYouTypeFormatter) InputDigit(nextChar rune) string {
	f.currentOutput = f.inputDigitWithOptionToRememberPosition(nextChar, false)
	return f.currentOutput
}

// InputDigitAndRememberPosition is like InputDigit but tracks the cursor position
// of this specific character through subsequent reformatting.
func (f *AsYouTypeFormatter) InputDigitAndRememberPosition(nextChar rune) string {
	f.currentOutput = f.inputDigitWithOptionToRememberPosition(nextChar, true)
	return f.currentOutput
}

// GetRememberedPosition returns the current position of the last
// InputDigitAndRememberPosition character in the formatted output.
func (f *AsYouTypeFormatter) GetRememberedPosition() int {
	if !f.ableToFormat {
		return f.originalPosition
	}
	accruedInputIndex := 0
	currentOutputIndex := 0
	runes := []rune(f.accruedInputWithoutFormatting.String())
	outputRunes := []rune(f.currentOutput)
	for accruedInputIndex < f.positionToRemember && currentOutputIndex < len(outputRunes) {
		if runes[accruedInputIndex] == outputRunes[currentOutputIndex] {
			accruedInputIndex++
		}
		currentOutputIndex++
	}
	return currentOutputIndex
}

func (f *AsYouTypeFormatter) inputDigitWithOptionToRememberPosition(nextChar rune, rememberPosition bool) string {
	f.accruedInput.WriteRune(nextChar)
	if rememberPosition {
		f.originalPosition = runeCount(&f.accruedInput)
	}

	// Format on-the-fly only when each character is a digit or a leading plus sign.
	if !f.isDigitOrLeadingPlusSign(nextChar) {
		f.ableToFormat = false
		f.inputHasFormatting = true
	} else {
		nextChar = f.normalizeAndAccrueDigitsAndPlusSign(nextChar, rememberPosition)
	}

	if !f.ableToFormat {
		if f.inputHasFormatting {
			return f.accruedInput.String()
		} else if f.attemptToExtractIdd() {
			if f.attemptToExtractCountryCallingCode() {
				return f.attemptToChoosePatternWithPrefixExtracted()
			}
		} else if f.ableToExtractLongerNdd() {
			f.prefixBeforeNationalNumber.WriteByte(byte(separatorBeforeNationalNumber))
			return f.attemptToChoosePatternWithPrefixExtracted()
		}
		return f.accruedInput.String()
	}

	// Start formatting attempts when at least MIN_LEADING_DIGITS_LENGTH digits have been entered.
	accruedLen := runeCount(&f.accruedInputWithoutFormatting)
	switch {
	case accruedLen <= 2:
		return f.accruedInput.String()
	case accruedLen == 3:
		if f.attemptToExtractIdd() {
			f.isExpectingCountryCallingCode = true
		} else {
			f.extractedNationalPrefix = f.removeNationalPrefixFromNationalNumber()
			return f.attemptToChooseFormattingPattern()
		}
		fallthrough
	default:
		if f.isExpectingCountryCallingCode {
			if f.attemptToExtractCountryCallingCode() {
				f.isExpectingCountryCallingCode = false
			}
			return f.prefixBeforeNationalNumber.String() + f.nationalNumber.String()
		}
		if len(f.possibleFormats) > 0 {
			tempNationalNumber := f.inputDigitHelper(nextChar)
			formattedNumber := f.attemptToFormatAccruedDigits()
			if formattedNumber != "" {
				return formattedNumber
			}
			f.narrowDownPossibleFormats(f.nationalNumber.String())
			if f.maybeCreateNewTemplate() {
				return f.inputAccruedNationalNumber()
			}
			if f.ableToFormat {
				return f.appendNationalNumber(tempNationalNumber)
			}
			return f.accruedInput.String()
		}
		return f.attemptToChooseFormattingPattern()
	}
}

func (f *AsYouTypeFormatter) attemptToChoosePatternWithPrefixExtracted() string {
	f.ableToFormat = true
	f.isExpectingCountryCallingCode = false
	f.possibleFormats = f.possibleFormats[:0]
	f.lastMatchPosition = 0
	f.formattingTemplate = f.formattingTemplate[:0]
	f.currentFormattingPattern = ""
	return f.attemptToChooseFormattingPattern()
}

func (f *AsYouTypeFormatter) isDigitOrLeadingPlusSign(nextChar rune) bool {
	if unicode.IsDigit(nextChar) {
		return true
	}
	return runeCount(&f.accruedInput) == 1 &&
		plusCharsPattern.MatchString(string(nextChar))
}

func (f *AsYouTypeFormatter) normalizeAndAccrueDigitsAndPlusSign(nextChar rune, rememberPosition bool) rune {
	var normalizedChar rune
	if nextChar == '+' || nextChar == '＋' {
		normalizedChar = '+'
		f.accruedInputWithoutFormatting.WriteRune(nextChar)
	} else {
		d, _ := unicodeDigitToASCII(nextChar)
		normalizedChar = rune(d)
		f.accruedInputWithoutFormatting.WriteByte(d)
		f.nationalNumber.WriteByte(d)
	}
	if rememberPosition {
		f.positionToRemember = runeCount(&f.accruedInputWithoutFormatting)
	}
	return normalizedChar
}

// attemptToExtractIdd extracts IDD and plus sign to prefixBeforeNationalNumber.
func (f *AsYouTypeFormatter) attemptToExtractIdd() bool {
	intlPrefix := f.currentMetadata.InternationalPrefix
	if intlPrefix == "" {
		return false
	}
	re, err := f.util.rc.getOrError(`\+|` + intlPrefix)
	if err != nil {
		return false
	}
	input := f.accruedInputWithoutFormatting.String()
	loc := re.FindStringIndex(input)
	if loc == nil || loc[0] != 0 {
		return false
	}
	f.isCompleteNumber = true
	startOfCC := loc[1]
	f.nationalNumber.Reset()
	f.nationalNumber.WriteString(input[startOfCC:])
	f.prefixBeforeNationalNumber.Reset()
	f.prefixBeforeNationalNumber.WriteString(input[:startOfCC])
	if input[0] != '+' {
		f.prefixBeforeNationalNumber.WriteByte(byte(separatorBeforeNationalNumber))
	}
	return true
}

// attemptToExtractCountryCallingCode extracts CC from nationalNumber.
func (f *AsYouTypeFormatter) attemptToExtractCountryCallingCode() bool {
	nn := f.nationalNumber.String()
	if nn == "" {
		return false
	}
	cc, remaining := f.util.extractCountryCode(nn)
	if cc == 0 {
		return false
	}
	f.nationalNumber.Reset()
	f.nationalNumber.WriteString(remaining)

	newRegionCode := f.util.getRegionCodeForCountryCode(cc)
	if newRegionCode == regionCodeNonGeo {
		f.currentMetadata = f.util.getMetadataForNonGeoRegion(cc)
	} else if newRegionCode != f.defaultCountry {
		f.currentMetadata = f.aytfGetMetadataForRegion(newRegionCode)
	}

	ccStr := strings.Builder{}
	ccStr.Grow(4)
	appendInt(&ccStr, int(cc))
	f.prefixBeforeNationalNumber.WriteString(ccStr.String())
	f.prefixBeforeNationalNumber.WriteByte(byte(separatorBeforeNationalNumber))
	// Clear previously extracted NDD since it's no longer valid.
	f.extractedNationalPrefix = ""
	return true
}

// removeNationalPrefixFromNationalNumber strips the national prefix and returns it.
func (f *AsYouTypeFormatter) removeNationalPrefixFromNationalNumber() string {
	nn := f.nationalNumber.String()
	startOfNationalNumber := 0

	if f.isNanpaNumberWithNationalPrefix() {
		startOfNationalNumber = 1
		f.prefixBeforeNationalNumber.WriteByte('1')
		f.prefixBeforeNationalNumber.WriteByte(byte(separatorBeforeNationalNumber))
		f.isCompleteNumber = true
	} else if f.currentMetadata.NationalPrefixForParsing != "" {
		re, err := f.util.rc.getOrError(f.currentMetadata.NationalPrefixForParsing)
		if err == nil {
			loc := re.FindStringIndex(nn)
			if loc != nil && loc[0] == 0 && loc[1] > 0 {
				f.isCompleteNumber = true
				startOfNationalNumber = loc[1]
				f.prefixBeforeNationalNumber.WriteString(nn[:startOfNationalNumber])
			}
		}
	}

	nationalPrefix := nn[:startOfNationalNumber]
	f.nationalNumber.Reset()
	f.nationalNumber.WriteString(nn[startOfNationalNumber:])
	return nationalPrefix
}

// isNanpaNumberWithNationalPrefix returns true if NANPA number starts with 1[2-9].
func (f *AsYouTypeFormatter) isNanpaNumberWithNationalPrefix() bool {
	nn := f.nationalNumber.String()
	if f.currentMetadata.CountryCode != nanpaCountryCode || len(nn) < 2 {
		return false
	}
	return nn[0] == '1' && nn[1] != '0' && nn[1] != '1'
}

// ableToExtractLongerNdd tries to extract a longer national prefix.
func (f *AsYouTypeFormatter) ableToExtractLongerNdd() bool {
	if f.extractedNationalPrefix != "" {
		// Put the extracted NDD back.
		nn := f.nationalNumber.String()
		f.nationalNumber.Reset()
		f.nationalNumber.WriteString(f.extractedNationalPrefix)
		f.nationalNumber.WriteString(nn)
		// Remove the previously extracted NDD from prefix.
		prefix := f.prefixBeforeNationalNumber.String()
		idx := strings.LastIndex(prefix, f.extractedNationalPrefix)
		if idx >= 0 {
			f.prefixBeforeNationalNumber.Reset()
			f.prefixBeforeNationalNumber.WriteString(prefix[:idx])
		}
	}
	newNP := f.removeNationalPrefixFromNationalNumber()
	return newNP != f.extractedNationalPrefix
}

// getAvailableFormats populates possibleFormats with eligible formats.
func (f *AsYouTypeFormatter) getAvailableFormats(leadingDigits string) {
	isInternationalNumber := f.isCompleteNumber && f.extractedNationalPrefix == ""
	var formatList []metadata.NumberFormat
	if isInternationalNumber && len(f.currentMetadata.IntlNumberFormats) > 0 {
		formatList = f.currentMetadata.IntlNumberFormats
	} else {
		formatList = f.currentMetadata.NumberFormats
	}

	for i := range formatList {
		format := &formatList[i]
		if f.extractedNationalPrefix != "" &&
			formattingRuleHasFirstGroupOnly(format.NationalPrefixFormattingRule) &&
			!format.NationalPrefixOptional &&
			format.DomesticCarrierCodeFormattingRule == "" {
			continue
		}
		if f.extractedNationalPrefix == "" &&
			!f.isCompleteNumber &&
			!formattingRuleHasFirstGroupOnly(format.NationalPrefixFormattingRule) &&
			!format.NationalPrefixOptional {
			continue
		}
		if eligibleFormatPattern.MatchString(format.Format) {
			f.possibleFormats = append(f.possibleFormats, format)
		}
	}
	f.narrowDownPossibleFormats(leadingDigits)
}

// narrowDownPossibleFormats removes formats that don't match leading digits.
func (f *AsYouTypeFormatter) narrowDownPossibleFormats(leadingDigits string) {
	indexOfLeadingDigitsPattern := len(leadingDigits) - minLeadingDigitsLength
	n := 0
	for _, format := range f.possibleFormats {
		if len(format.LeadingDigitsPatterns) == 0 {
			// Keep formats without leading digits restrictions.
			f.possibleFormats[n] = format
			n++
			continue
		}
		lastIdx := indexOfLeadingDigitsPattern
		if lastIdx >= len(format.LeadingDigitsPatterns) {
			lastIdx = len(format.LeadingDigitsPatterns) - 1
		}
		pattern := format.LeadingDigitsPatterns[lastIdx]
		re, err := f.util.rc.getOrError(`^(?:` + pattern + `)`)
		if err != nil {
			continue
		}
		if re.MatchString(leadingDigits) {
			f.possibleFormats[n] = format
			n++
		}
	}
	f.possibleFormats = f.possibleFormats[:n]
}

// maybeCreateNewTemplate attempts to create a new formatting template.
// Returns true if a new template is created (as opposed to reusing existing).
func (f *AsYouTypeFormatter) maybeCreateNewTemplate() bool {
	for i := 0; i < len(f.possibleFormats); {
		format := f.possibleFormats[i]
		pattern := format.Pattern
		if f.currentFormattingPattern == pattern {
			return false
		}
		if f.createFormattingTemplate(format) {
			f.currentFormattingPattern = pattern
			f.shouldAddSpaceAfterNationalPrefix =
				nationalPrefixSeparatorsPattern.MatchString(format.NationalPrefixFormattingRule)
			f.lastMatchPosition = 0
			return true
		}
		// Remove this format from the list.
		f.possibleFormats = append(f.possibleFormats[:i], f.possibleFormats[i+1:]...)
	}
	f.ableToFormat = false
	return false
}

// createFormattingTemplate creates a template from the given format.
func (f *AsYouTypeFormatter) createFormattingTemplate(format *metadata.NumberFormat) bool {
	numberPattern := format.Pattern
	f.formattingTemplate = f.formattingTemplate[:0]
	template := f.getFormattingTemplate(numberPattern, format.Format)
	if len(template) > 0 {
		f.formattingTemplate = template
		return true
	}
	return false
}

// getFormattingTemplate creates a template by matching "999..." against the pattern.
func (f *AsYouTypeFormatter) getFormattingTemplate(numberPattern, numberFormat string) []rune {
	const longestPhoneNumber = "999999999999999"
	re, err := f.util.rc.getOrError(numberPattern)
	if err != nil {
		return nil
	}
	match := re.FindString(longestPhoneNumber)
	if match == "" {
		return nil
	}
	// If the matched number is shorter than what we've accumulated, template is invalid.
	if len(match) < runeCount(&f.nationalNumber) {
		return nil
	}
	// Format the matched number.
	template := re.ReplaceAllString(match, numberFormat)
	// Replace each '9' with the digit placeholder.
	runes := []rune(template)
	for i, r := range runes {
		if r == '9' {
			runes[i] = digitPlaceholder
		}
	}
	return runes
}

// inputDigitHelper fills the next placeholder in the template.
func (f *AsYouTypeFormatter) inputDigitHelper(nextChar rune) string {
	for i := f.lastMatchPosition; i < len(f.formattingTemplate); i++ {
		if f.formattingTemplate[i] == digitPlaceholder {
			f.formattingTemplate[i] = nextChar
			f.lastMatchPosition = i
			return string(f.formattingTemplate[:i+1])
		}
	}
	// No more placeholders — template is full.
	if len(f.possibleFormats) == 1 {
		f.ableToFormat = false
	}
	f.currentFormattingPattern = ""
	return f.accruedInput.String()
}

// attemptToFormatAccruedDigits tries to do an exact format match on the current digits.
func (f *AsYouTypeFormatter) attemptToFormatAccruedDigits() string {
	nn := f.nationalNumber.String()
	for _, format := range f.possibleFormats {
		re, err := f.util.rc.getOrError(format.Pattern)
		if err != nil {
			continue
		}
		if !f.util.matchesEntirely(format.Pattern, nn) {
			continue
		}
		f.shouldAddSpaceAfterNationalPrefix =
			nationalPrefixSeparatorsPattern.MatchString(format.NationalPrefixFormattingRule)
		formattedNumber := re.ReplaceAllString(nn, format.Format)
		fullOutput := f.appendNationalNumber(formattedNumber)
		// Check that formatting didn't add or remove digits.
		formattedDigitsOnly := NormalizeDiallableCharsOnly(fullOutput)
		if formattedDigitsOnly == f.accruedInputWithoutFormatting.String() {
			return fullOutput
		}
	}
	return ""
}

// attemptToChooseFormattingPattern selects initial formatting pattern.
func (f *AsYouTypeFormatter) attemptToChooseFormattingPattern() string {
	nn := f.nationalNumber.String()
	if len(nn) >= minLeadingDigitsLength {
		f.getAvailableFormats(nn)
		formattedNumber := f.attemptToFormatAccruedDigits()
		if formattedNumber != "" {
			return formattedNumber
		}
		if f.maybeCreateNewTemplate() {
			return f.inputAccruedNationalNumber()
		}
		return f.accruedInput.String()
	}
	return f.appendNationalNumber(nn)
}

// inputAccruedNationalNumber re-feeds all national digits through the template.
func (f *AsYouTypeFormatter) inputAccruedNationalNumber() string {
	nn := f.nationalNumber.String()
	if len(nn) > 0 {
		var tempNationalNumber string
		for _, ch := range nn {
			tempNationalNumber = f.inputDigitHelper(ch)
		}
		if f.ableToFormat {
			return f.appendNationalNumber(tempNationalNumber)
		}
		return f.accruedInput.String()
	}
	return f.prefixBeforeNationalNumber.String()
}

// appendNationalNumber combines the prefix with the national number.
func (f *AsYouTypeFormatter) appendNationalNumber(nationalNumber string) string {
	prefix := f.prefixBeforeNationalNumber.String()
	prefixLen := len(prefix)
	if f.shouldAddSpaceAfterNationalPrefix && prefixLen > 0 &&
		prefix[prefixLen-1] != byte(separatorBeforeNationalNumber) {
		return prefix + string(separatorBeforeNationalNumber) + nationalNumber
	}
	return prefix + nationalNumber
}

// formattingRuleHasFirstGroupOnly returns true if the formatting rule only
// references $1 (meaning national prefix is effectively optional).
func formattingRuleHasFirstGroupOnly(rule string) bool {
	return rule == "" || firstGroupOnlyPrefixPattern.MatchString(rule)
}

// --- Helpers ---

// runeCount returns the number of runes in a strings.Builder.
func runeCount(b *strings.Builder) int {
	return len([]rune(b.String()))
}

// appendInt appends an integer as decimal digits to a strings.Builder.
func appendInt(b *strings.Builder, n int) {
	if n == 0 {
		b.WriteByte('0')
		return
	}
	// Compute digits.
	var digits [10]byte
	i := len(digits)
	for n > 0 {
		i--
		digits[i] = byte('0' + n%10)
		n /= 10
	}
	b.Write(digits[i:])
}
