package phonesafe

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/shepard-labs/go-phonesafe/internal/metadata"
)

// --- Matcher-specific pre-compiled patterns ---

var (
	// matcherPattern is the main candidate detection pattern used by the matcher.
	// Similar to validPhoneNumberPattern but bounded for scanning within text.
	matcherPattern *regexp.Regexp

	// matchingBracketsPattern ensures brackets in a candidate are balanced.
	matchingBracketsPattern *regexp.Regexp

	// pubPagesPattern matches publication page ranges (e.g., "211-227 (2003)").
	pubPagesPattern *regexp.Regexp

	// slashSeparatedDatesPattern matches dates like "3/10/2011" or "31/10/96".
	slashSeparatedDatesPattern *regexp.Regexp

	// timeStampsPattern matches timestamps like "2012-01-02 08".
	timeStampsPattern *regexp.Regexp

	// timeStampsSuffixPattern matches the ":mm" suffix of a timestamp.
	timeStampsSuffixPattern *regexp.Regexp

	// leadClassPattern matches opening phone-number punctuation (bracket or plus).
	leadClassPattern *regexp.Regexp

	// innerMatchPatterns are ordered patterns to extract phone numbers from a
	// larger candidate. Ordered by specificity.
	innerMatchPatterns [6]*regexp.Regexp
)

func init() {
	// Build the extension pattern for matching (no auto-dialling patterns).
	extnForMatching := buildExtnPatternForMatching()

	// Character classes matching upstream.
	const openingParens = `(\[\x{FF08}\x{FF3B}`
	const closingParens = `)\]\x{FF09}\x{FF3D}`
	nonParens := `[^` + openingParens + closingParens + `]`

	// Matching brackets: allows optional unmatched opening bracket, then content
	// with up to 3 matched pairs.
	matchingBracketsPattern = regexp.MustCompile(
		`^(?:[` + openingParens + `])?` +
			`(?:` + nonParens + `+[` + closingParens + `])?` +
			nonParens + `+` +
			`(?:[` + openingParens + `]` + nonParens + `+[` + closingParens + `]){0,3}` +
			nonParens + `*$`)

	// Lead class: opening brackets and plus chars.
	const plusChars = `+\x{FF0B}`
	leadClassChars := openingParens + plusChars
	leadClassPattern = regexp.MustCompile(`[` + leadClassChars + `]`)

	// Valid punctuation for the matcher pattern (same as parse.go but written explicitly).
	const validPunct = `\x{2010}-\x{2015}\x{2212}\x{30FC}\x{FF0D}-\x{FF0F}` +
		` \x{00A0}\x{00AD}\x{200B}\x{2060}\x{3000}` +
		`()\x{FF08}\x{FF09}\x{FF3B}\x{FF3D}\x{2053}\x{223C}` +
		`x.\\\/~\-`

	const digits = `\p{Nd}`
	digitBlockLimit := maxLengthForNSN + maxLengthCC // 17 + 3 = 20

	punctuation := `[` + validPunct + `]{0,4}`
	digitSequence := `[` + digits + `]{1,` + strconv.Itoa(digitBlockLimit) + `}`
	blockLimit := `{0,` + strconv.Itoa(digitBlockLimit) + `}`
	leadLimit := `{0,2}`

	leadClass := `[` + leadClassChars + `]`

	// Main matcher pattern: lead chars + punctuation limited, digit sequences bounded.
	matcherPattern = regexp.MustCompile(
		`(?i)(?:` + leadClass + punctuation + `)` + leadLimit +
			digitSequence + `(?:` + punctuation + digitSequence + `)` + blockLimit +
			`(?:` + extnForMatching + `)?`)

	// Filter patterns.
	pubPagesPattern = regexp.MustCompile(`\d{1,5}-+\d{1,5}\s{0,4}\(\d{1,4}`)
	slashSeparatedDatesPattern = regexp.MustCompile(
		`(?:(?:[0-3]?\d/[01]?\d)|(?:[01]?\d/[0-3]?\d))/(?:[12]\d)?\d{2}`)
	timeStampsPattern = regexp.MustCompile(`[12]\d{3}[-/]?[01]\d[-/]?[0-3]\d +[0-2]\d$`)
	timeStampsSuffixPattern = regexp.MustCompile(`^:[0-5]\d`)

	// Inner match patterns (ordered by specificity).
	innerMatchPatterns = [6]*regexp.Regexp{
		regexp.MustCompile(`/+(.*)`),
		regexp.MustCompile(`(\([^(]*)`),
		regexp.MustCompile(`(?:\p{Z}-|-\p{Z})\p{Z}*(.+)`),
		regexp.MustCompile(`[\x{2012}-\x{2015}\x{FF0D}]\p{Z}*(.+)`),
		regexp.MustCompile(`\.+\p{Z}*([^.]+)`),
		regexp.MustCompile(`\p{Z}+(\P{Z}+)`),
	}
}

// buildExtnPatternForMatching builds the extension regex for matching context
// (excludes auto-dialling patterns that are only valid for parsing).
func buildExtnPatternForMatching() string {
	const digits = `\p{Nd}`

	extnDigits := func(maxLen int) string {
		return `([` + digits + `]{1,` + strconv.Itoa(maxLen) + `})`
	}

	separatorBetween := `[ \x{00A0}\t,]*`
	charsAfterLabel := `[:\.\x{FF0E}]?[ \x{00A0}\t,\-]*`
	optionalSuffix := `#?`

	explicitLabels := `(?:e?xt(?:ensi(?:o\x{0301}?|\x{00F3}))?n?|` +
		`(?:\x{FF45})?\x{FF58}\x{FF54}(?:\x{FF4E})?|` +
		`\x{0434}\x{043E}\x{0431}|anexo)`
	ambiguousLabels := `(?:[x\x{FF58}#\x{FF03}~\x{FF5E}]|int|\x{FF49}\x{FF4E}\x{FF54})`
	ambiguousSeparator := `[- ]+`

	rfcExtn := rfc3966ExtnPrefix + extnDigits(20)
	explicitExtn := separatorBetween + explicitLabels + charsAfterLabel + extnDigits(20) + optionalSuffix
	ambiguousExtn := separatorBetween + ambiguousLabels + charsAfterLabel + extnDigits(9) + optionalSuffix
	americanExtn := ambiguousSeparator + extnDigits(6) + `#`

	return `(?i)(?:` + rfcExtn + `|` + explicitExtn + `|` + ambiguousExtn + `|` + americanExtn + `)`
}

// --- Utility functions ---

// normalizeDigitsKeepNonDigits normalizes Unicode digits to ASCII while keeping
// non-digit characters in place. Used for grouping validation.
func normalizeDigitsKeepNonDigits(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if d, ok := unicodeDigitToASCII(r); ok {
			b.WriteByte(d)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// isLatinLetter returns true if the rune is a Latin-script letter or combining
// diacritical mark.
func isLatinLetter(r rune) bool {
	if !unicode.IsLetter(r) && !unicode.Is(unicode.Mn, r) {
		return false
	}
	// Check if the rune falls in Latin blocks or combining marks.
	return (r >= 0x0041 && r <= 0x007A) || // Basic Latin A-z
		(r >= 0x00C0 && r <= 0x00FF) || // Latin-1 Supplement
		(r >= 0x0100 && r <= 0x017F) || // Latin Extended-A
		(r >= 0x0180 && r <= 0x024F) || // Latin Extended-B
		(r >= 0x1E00 && r <= 0x1EFF) || // Latin Extended Additional
		(r >= 0x0300 && r <= 0x036F) // Combining Diacritical Marks
}

// isInvalidPunctuationSymbol returns true for characters that shouldn't appear
// adjacent to a phone number (% and currency symbols).
func isInvalidPunctuationSymbol(r rune) bool {
	return r == '%' || unicode.Is(unicode.Sc, r)
}

// containsOnlyValidXChars validates that 'x'/'X' characters in the candidate
// are either carrier codes (xx before NSN) or extension separators.
func containsOnlyValidXChars(number PhoneNumber, candidate string, util *PhoneNumberUtil) bool {
	for i := 0; i < len(candidate)-1; i++ {
		c := candidate[i]
		if c == 'x' || c == 'X' {
			next := candidate[i+1]
			if next == 'x' || next == 'X' {
				// Carrier code case: two consecutive x's before NSN.
				i++
				parsed, err := util.Parse(candidate[i:], "ZZ")
				if err != nil {
					return false
				}
				if util.IsNumberMatch(number, parsed) != MatchNSN {
					return false
				}
			} else {
				// Extension sign case: digits after x should equal the extension.
				digitsAfterX := NormalizeDigitsOnly(candidate[i:])
				if digitsAfterX != number.Extension {
					return false
				}
			}
		}
	}
	return true
}

// containsMoreThanOneSlashInNationalNumber returns true if there are multiple
// slashes within the national portion of the number candidate.
func containsMoreThanOneSlashInNationalNumber(number PhoneNumber, candidate string) bool {
	firstSlash := strings.IndexByte(candidate, '/')
	if firstSlash < 0 {
		return false
	}
	secondSlash := strings.IndexByte(candidate[firstSlash+1:], '/')
	if secondSlash < 0 {
		return false
	}
	secondSlash += firstSlash + 1

	// If the first slash is after the country code, check for more slashes.
	candidateHasCC := number.CountryCodeSource == CountryCodeFromNumberWithPlus ||
		number.CountryCodeSource == CountryCodeFromNumberWithoutPlus
	if candidateHasCC {
		beforeSlash := NormalizeDigitsOnly(candidate[:firstSlash])
		ccStr := strconv.FormatInt(int64(number.CountryCode), 10)
		if beforeSlash == ccStr {
			// First slash was after country code; any more slashes are invalid.
			return strings.Contains(candidate[secondSlash+1:], "/")
		}
	}
	return true
}

// isNationalPrefixPresentIfRequired checks that the national prefix is present
// in the raw input when it's required by the formatting rules.
func isNationalPrefixPresentIfRequired(number PhoneNumber, util *PhoneNumberUtil) bool {
	// If the country code came from the international format, prefix is not required.
	if number.CountryCodeSource != CountryCodeFromDefaultCountry {
		return true
	}

	region := util.getRegionCodeForCountryCode(number.CountryCode)
	meta := util.getMetadataForRegion(region)
	if meta == nil {
		return true
	}

	nationalNumber := GetNationalSignificantNumber(number)
	formatRule := util.chooseFormattingPatternForNumber(meta.NumberFormats, nationalNumber)
	if formatRule == nil || formatRule.NationalPrefixFormattingRule == "" {
		return true
	}

	if formatRule.NationalPrefixOptional {
		return true
	}
	if formattingRuleHasFirstGroupOnly(formatRule.NationalPrefixFormattingRule) {
		return true
	}

	// Check if raw input starts with the national prefix.
	rawInput := NormalizeDigitsOnly(number.RawInput)
	_, _, stripped := util.maybeStripNationalPrefixAndCarrierCode(meta, rawInput)
	return stripped
}

// --- Grouping validation ---

// numberGroupingChecker is the function type for grouping validation.
type numberGroupingChecker func(util *PhoneNumberUtil, number PhoneNumber, normalizedCandidate string, expectedGroups []string) bool

// getNationalNumberGroups returns digit groups of the national number formatted
// per standard rules (by formatting as RFC3966 and splitting on '-').
func getNationalNumberGroups(util *PhoneNumberUtil, number PhoneNumber) []string {
	rfc := util.Format(number, FormatRFC3966)
	// Remove extension part.
	if idx := strings.Index(rfc, ";"); idx >= 0 {
		rfc = rfc[:idx]
	}
	// Skip "tel:+CC-" prefix — find first '-' after "tel:+"
	startIdx := strings.IndexByte(rfc, '-')
	if startIdx < 0 {
		return []string{rfc}
	}
	return strings.Split(rfc[startIdx+1:], "-")
}

// getNationalNumberGroupsForPattern returns digit groups when formatting with
// a specific pattern (for alternate format checking).
func getNationalNumberGroupsForPattern(util *PhoneNumberUtil, number PhoneNumber, pattern *metadata.NumberFormat) []string {
	nsn := GetNationalSignificantNumber(number)
	formatted := util.formatNsnUsingPatternWithCarrier(nsn, pattern, FormatRFC3966, "", "")
	return strings.Split(formatted, "-")
}

// allNumberGroupsRemainGrouped checks that expected digit groups appear
// consecutively in the normalized candidate.
func allNumberGroupsRemainGrouped(util *PhoneNumberUtil, number PhoneNumber, normalizedCandidate string, formattedGroups []string) bool {
	fromIndex := 0
	if number.CountryCodeSource != CountryCodeFromDefaultCountry {
		// Skip past the country code in the candidate.
		ccStr := strconv.FormatInt(int64(number.CountryCode), 10)
		idx := strings.Index(normalizedCandidate, ccStr)
		if idx >= 0 {
			fromIndex = idx + len(ccStr)
		}
	}

	for i, group := range formattedGroups {
		idx := strings.Index(normalizedCandidate[fromIndex:], group)
		if idx < 0 {
			return false
		}
		fromIndex += idx + len(group)

		if i == 0 && fromIndex < len(normalizedCandidate) {
			// After NDC: if region has a national prefix and the next char is a digit,
			// then there's no formatting symbol after NDC. In that case we require
			// the entire NSN to be unformatted (single block).
			region := util.getRegionCodeForCountryCode(number.CountryCode)
			nddPrefix := util.getNddPrefixForRegion(region, true)
			if nddPrefix != "" {
				r, _ := utf8.DecodeRuneInString(normalizedCandidate[fromIndex:])
				if r >= '0' && r <= '9' {
					nsn := GetNationalSignificantNumber(number)
					return strings.HasPrefix(normalizedCandidate[fromIndex-len(group):], nsn)
				}
			}
		}
	}

	// Verify extension hasn't been accidentally consumed as the last group.
	remainder := normalizedCandidate[fromIndex:]
	return strings.Contains(remainder, number.Extension)
}

// allNumberGroupsAreExactlyPresent checks that candidate digit groups exactly
// match the expected formatting groups.
func allNumberGroupsAreExactlyPresent(util *PhoneNumberUtil, number PhoneNumber, normalizedCandidate string, formattedGroups []string) bool {
	// Split candidate into digit groups.
	candidateGroups := splitOnNonDigits(normalizedCandidate)

	// Determine the last group index (skip extension group if present).
	candidateIdx := len(candidateGroups) - 1
	if number.Extension != "" && candidateIdx > 0 {
		candidateIdx--
	}

	// Check if NSN is a single block.
	if len(candidateGroups) == 1 ||
		(candidateIdx >= 0 && strings.Contains(candidateGroups[candidateIdx], GetNationalSignificantNumber(number))) {
		return true
	}

	// Compare from end, skipping the first formatted group.
	formattedIdx := len(formattedGroups) - 1
	for formattedIdx > 0 && candidateIdx >= 0 {
		if candidateGroups[candidateIdx] != formattedGroups[formattedIdx] {
			return false
		}
		formattedIdx--
		candidateIdx--
	}

	// First group: candidate group should end with the expected first group
	// (may have national prefix prepended).
	return candidateIdx >= 0 && strings.HasSuffix(candidateGroups[candidateIdx], formattedGroups[0])
}

// splitOnNonDigits splits a string into groups of consecutive digits.
func splitOnNonDigits(s string) []string {
	var groups []string
	var current strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				groups = append(groups, current.String())
				current.Reset()
			}
		}
	}
	if current.Len() > 0 {
		groups = append(groups, current.String())
	}
	return groups
}

// checkNumberGroupingIsValid checks the candidate against primary and alternate
// formatting patterns using the provided checker function.
func (m *PhoneNumberMatcher) checkNumberGroupingIsValid(
	number PhoneNumber, candidate string, util *PhoneNumberUtil, checker numberGroupingChecker) bool {

	normalizedCandidate := normalizeDigitsKeepNonDigits(candidate)
	formattedGroups := getNationalNumberGroups(util, number)
	if checker(util, number, normalizedCandidate, formattedGroups) {
		return true
	}

	// Try alternate formats.
	altFormats, ok := metadata.AlternateFormats[number.CountryCode]
	if !ok {
		return false
	}

	nsn := GetNationalSignificantNumber(number)
	for i := range altFormats {
		af := &altFormats[i]
		if len(af.LeadingDigitsPatterns) > 0 {
			leadingDigitsPattern := af.LeadingDigitsPatterns[0]
			re, err := util.rc.getOrError(`^(?:` + leadingDigitsPattern + `)`)
			if err != nil || !re.MatchString(nsn) {
				continue
			}
		}
		groups := getNationalNumberGroupsForPattern(util, number, af)
		if checker(util, number, normalizedCandidate, groups) {
			return true
		}
	}
	return false
}

// --- Leniency verification ---

// verifyLeniency checks a parsed number against the specified leniency level.
func verifyLeniency(leniency Leniency, number PhoneNumber, candidate string, util *PhoneNumberUtil, m *PhoneNumberMatcher) bool {
	switch leniency {
	case LeniencyPossible:
		return util.IsPossibleNumber(number)

	case LeniencyValid:
		if !util.IsValidNumber(number) {
			return false
		}
		if !containsOnlyValidXChars(number, candidate, util) {
			return false
		}
		return isNationalPrefixPresentIfRequired(number, util)

	case LeniencyStrictGrouping:
		if !util.IsValidNumber(number) {
			return false
		}
		if !containsOnlyValidXChars(number, candidate, util) {
			return false
		}
		if containsMoreThanOneSlashInNationalNumber(number, candidate) {
			return false
		}
		if !isNationalPrefixPresentIfRequired(number, util) {
			return false
		}
		return m.checkNumberGroupingIsValid(number, candidate, util, allNumberGroupsRemainGrouped)

	case LeniencyExactGrouping:
		if !util.IsValidNumber(number) {
			return false
		}
		if !containsOnlyValidXChars(number, candidate, util) {
			return false
		}
		if containsMoreThanOneSlashInNationalNumber(number, candidate) {
			return false
		}
		if !isNationalPrefixPresentIfRequired(number, util) {
			return false
		}
		return m.checkNumberGroupingIsValid(number, candidate, util, allNumberGroupsAreExactlyPresent)
	}
	return false
}

// --- PhoneNumberMatcher ---

// matcherState represents the iteration state.
type matcherState int

const (
	matcherNotReady matcherState = iota
	matcherReady
	matcherDone
)

// PhoneNumberMatcher finds and extracts phone numbers from text. It is not
// goroutine-safe. Use FindNumbers to create instances.
type PhoneNumberMatcher struct {
	util            *PhoneNumberUtil
	text            string
	preferredRegion string
	leniency        Leniency
	maxTries        int64

	state       matcherState
	lastMatch   *PhoneNumberMatch
	searchIndex int
}

// FindNumbers creates a PhoneNumberMatcher that extracts phone numbers from
// the given text. The region parameter specifies the default region for numbers
// without international prefix. Use "ZZ" for numbers that must have a leading "+".
// The maxTries parameter limits attempts on invalid candidates (use 65535 for typical use).
func (u *PhoneNumberUtil) FindNumbers(text, region string, leniency Leniency, maxTries int) *PhoneNumberMatcher {
	if maxTries < 0 {
		maxTries = 0
	}
	return &PhoneNumberMatcher{
		util:            u,
		text:            text,
		preferredRegion: region,
		leniency:        leniency,
		maxTries:        int64(maxTries),
		state:           matcherNotReady,
	}
}

// HasNext returns true if there is another phone number match in the text.
// Advances internal state by searching for the next match if needed.
func (m *PhoneNumberMatcher) HasNext() bool {
	if m.state == matcherNotReady {
		match := m.find(m.searchIndex)
		if match == nil {
			m.state = matcherDone
		} else {
			m.lastMatch = match
			m.searchIndex = match.End
			m.state = matcherReady
		}
	}
	return m.state == matcherReady
}

// Next returns the next phone number match. Must be called after HasNext()
// returns true. Panics if called when no match is available.
func (m *PhoneNumberMatcher) Next() PhoneNumberMatch {
	if !m.HasNext() {
		panic("phonesafe: PhoneNumberMatcher.Next called with no remaining matches")
	}
	result := *m.lastMatch
	m.lastMatch = nil
	m.state = matcherNotReady
	return result
}

// find searches for the next phone number candidate starting at index.
func (m *PhoneNumberMatcher) find(index int) *PhoneNumberMatch {
	for m.maxTries > 0 {
		loc := matcherPattern.FindStringIndex(m.text[index:])
		if loc == nil {
			break
		}
		start := index + loc[0]
		candidate := m.text[start : index+loc[1]]

		// Trim trailing content that looks like a second number.
		candidate = trimAfterFirstMatchStr(secondNumberStartPattern, candidate)

		match := m.extractMatch(candidate, start)
		if match != nil {
			return match
		}

		index = start + len(candidate)
		m.maxTries--
	}
	return nil
}

// trimAfterFirstMatchStr trims the candidate after the first match of pattern.
func trimAfterFirstMatchStr(pattern *regexp.Regexp, candidate string) string {
	loc := pattern.FindStringIndex(candidate)
	if loc != nil {
		return candidate[:loc[0]]
	}
	return candidate
}

// extractMatch attempts to extract a phone number match from a candidate.
func (m *PhoneNumberMatcher) extractMatch(candidate string, offset int) *PhoneNumberMatch {
	// Skip dates.
	if slashSeparatedDatesPattern.MatchString(candidate) {
		return nil
	}

	// Skip timestamps.
	if timeStampsPattern.MatchString(candidate) {
		followingText := ""
		endIdx := offset + len(candidate)
		if endIdx < len(m.text) {
			followingText = m.text[endIdx:]
		}
		if timeStampsSuffixPattern.MatchString(followingText) {
			return nil
		}
	}

	// Try the full candidate.
	match := m.parseAndVerify(candidate, offset)
	if match != nil {
		return match
	}

	// Try inner matches.
	return m.extractInnerMatch(candidate, offset)
}

// extractInnerMatch tries to extract phone numbers from within a larger candidate.
func (m *PhoneNumberMatcher) extractInnerMatch(candidate string, offset int) *PhoneNumberMatch {
	for _, pattern := range innerMatchPatterns {
		allMatches := pattern.FindAllStringSubmatchIndex(candidate, -1)
		isFirstMatch := true

		for _, loc := range allMatches {
			if m.maxTries <= 0 {
				return nil
			}

			if isFirstMatch {
				// Try text before the first inner match.
				groupBefore := candidate[:loc[0]]
				groupBefore = trimAfterFirstMatchStr(unwantedEndCharPattern, groupBefore)
				match := m.parseAndVerify(groupBefore, offset)
				if match != nil {
					return match
				}
				m.maxTries--
				isFirstMatch = false
			}

			// Try the captured group (submatch index 2,3 is group 1).
			if loc[2] >= 0 && loc[3] >= 0 {
				group := candidate[loc[2]:loc[3]]
				group = trimAfterFirstMatchStr(unwantedEndCharPattern, group)
				match := m.parseAndVerify(group, offset+loc[2])
				if match != nil {
					return match
				}
				m.maxTries--
			}
		}
	}
	return nil
}

// parseAndVerify parses the candidate and verifies it against the leniency level.
func (m *PhoneNumberMatcher) parseAndVerify(candidate string, offset int) *PhoneNumberMatch {
	if candidate == "" {
		return nil
	}

	// Check bracket matching.
	if !matchingBracketsPattern.MatchString(candidate) {
		return nil
	}

	// Check for publication pages.
	if pubPagesPattern.MatchString(candidate) {
		return nil
	}

	// For VALID and stricter: check surrounding characters.
	if m.leniency >= LeniencyValid {
		// Check character before candidate.
		if offset > 0 && !leadClassPattern.MatchString(candidate[:1]) {
			prevRune, _ := utf8.DecodeLastRuneInString(m.text[:offset])
			if prevRune != utf8.RuneError {
				if isInvalidPunctuationSymbol(prevRune) || isLatinLetter(prevRune) {
					return nil
				}
			}
		}
		// Check character after candidate.
		endIdx := offset + len(candidate)
		if endIdx < len(m.text) {
			nextRune, _ := utf8.DecodeRuneInString(m.text[endIdx:])
			if nextRune != utf8.RuneError {
				if isInvalidPunctuationSymbol(nextRune) || isLatinLetter(nextRune) {
					return nil
				}
			}
		}
	}

	// Parse the candidate.
	number, err := m.util.ParseAndKeepRawInput(candidate, m.preferredRegion)
	if err != nil {
		return nil
	}

	// Verify against leniency.
	if !verifyLeniency(m.leniency, number, candidate, m.util, m) {
		return nil
	}

	// Clear extra fields (match only stores core number identity).
	number.CountryCodeSource = CountryCodeUnspecified
	number.RawInput = ""
	number.PreferredDomesticCarrierCode = ""

	return &PhoneNumberMatch{
		Start:     offset,
		End:       offset + len(candidate),
		RawString: candidate,
		Number:    number,
	}
}
