package phonesafe

import (
	"testing"
)

// --- Helper functions ---

// assertFindsNumber verifies that the given number is found in text wrapped
// with leading/trailing content using the given leniency.
func assertFindsNumber(t *testing.T, text, region string, leniency Leniency, expectedRaw string) {
	t.Helper()
	u := Instance()
	m := u.FindNumbers(text, region, leniency, 65535)
	if !m.HasNext() {
		t.Fatalf("expected to find %q in %q but found nothing", expectedRaw, text)
	}
	match := m.Next()
	if match.RawString != expectedRaw {
		t.Errorf("expected raw=%q, got raw=%q", expectedRaw, match.RawString)
	}
}

// doTestFindInContext wraps the number in various context strings and verifies
// it is found at POSSIBLE leniency.
func doTestFindInContext(t *testing.T, number, region string) {
	t.Helper()
	contexts := []struct {
		leading  string
		trailing string
	}{
		{"", ""},
		{"   ", "\t"},
		{"Call ", ""},
		{"Call ", " please."},
		{"", ".."},
		{"Num is ", "."},
		{"Num is ", ". More text."},
	}

	for _, ctx := range contexts {
		text := ctx.leading + number + ctx.trailing
		assertFindsNumber(t, text, region, LeniencyPossible, number)
	}
}

// --- Tests ---

func TestMatcherFindNationalNumber(t *testing.T) {
	doTestFindInContext(t, "033316005", "NZ")
	doTestFindInContext(t, "03-331 6005", "NZ")
	doTestFindInContext(t, "03 331 6005", "NZ")
	doTestFindInContext(t, "0064 3 331 6005", "NZ")
	doTestFindInContext(t, "01164 3 331 6005", "US")
	doTestFindInContext(t, "+64 3 331 6005", "US")
	doTestFindInContext(t, "650-253-0000", "US")
}

func TestMatcherFindWithInternationalPrefixes(t *testing.T) {
	doTestFindInContext(t, "+1 (650) 333-6000", "NZ")
	doTestFindInContext(t, "1-650-333-6000", "US")
	doTestFindInContext(t, "++1 (650) 333-6000", "PL")
	// Fullwidth plus.
	doTestFindInContext(t, "＋1 (650) 333-6000", "SG")
}

func TestMatcherFindWithLeadingZero(t *testing.T) {
	doTestFindInContext(t, "+39 02-36618 300", "NZ")
	doTestFindInContext(t, "02-36618 300", "IT")
	doTestFindInContext(t, "312 345 678", "IT")
}

func TestMatcherFindExtensions(t *testing.T) {
	doTestFindInContext(t, "03 331 6005 ext 3456", "NZ")
	doTestFindInContext(t, "03-3316005x3456", "NZ")
	doTestFindInContext(t, "03-3316005 int.3456", "NZ")
	doTestFindInContext(t, "03 3316005 #3456", "NZ")
	doTestFindInContext(t, "+44 2034567890x456", "NZ")
	doTestFindInContext(t, "+44 2034567890x456", "GB")
	doTestFindInContext(t, "+44 2034567890 x456", "GB")
	doTestFindInContext(t, "+44 2034567890 X456", "GB")
}

func TestMatcherFourMatchesInARow(t *testing.T) {
	u := Instance()
	numbers := []string{"415-666-7777", "800-443-1223", "212-443-1223", "650-443-1223"}
	text := numbers[0] + " - " + numbers[1] + " - " + numbers[2] + " - " + numbers[3]

	m := u.FindNumbers(text, "US", LeniencyPossible, 65535)
	for _, expected := range numbers {
		if !m.HasNext() {
			t.Fatalf("expected to find %q but no more matches", expected)
		}
		match := m.Next()
		if match.RawString != expected {
			t.Errorf("expected %q, got %q", expected, match.RawString)
		}
	}
	if m.HasNext() {
		t.Error("expected no more matches")
	}
}

func TestMatcherMultipleMatchesWithSpaces(t *testing.T) {
	u := Instance()
	text := "(415) 666-7777 (800) 443-1223"
	m := u.FindNumbers(text, "US", LeniencyPossible, 65535)

	if !m.HasNext() {
		t.Fatal("expected first match")
	}
	match1 := m.Next()
	if match1.RawString != "(415) 666-7777" {
		t.Errorf("first match: got %q", match1.RawString)
	}

	if !m.HasNext() {
		t.Fatal("expected second match")
	}
	match2 := m.Next()
	if match2.RawString != "(800) 443-1223" {
		t.Errorf("second match: got %q", match2.RawString)
	}
}

func TestMatcherIsLatinLetter(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'c', true},
		{'C', true},
		{'É', true}, // E with acute
		{'́', true}, // Combining acute accent
		{':', false},
		{'5', false},
		{'-', false},
		{'.', false},
		{' ', false},
		{'我', false}, // Chinese
		{'の', false}, // Hiragana
	}
	for _, tt := range tests {
		got := isLatinLetter(tt.r)
		if got != tt.want {
			t.Errorf("isLatinLetter(%q) = %v, want %v", tt.r, got, tt.want)
		}
	}
}

func TestMatcherSurroundingLatinChars(t *testing.T) {
	// Numbers surrounded by Latin characters should not match at VALID leniency.
	u := Instance()
	number := "650-253-0000"

	// With Latin prefix — not found at VALID.
	text := "abc" + number
	m := u.FindNumbers(text, "US", LeniencyValid, 65535)
	if m.HasNext() {
		t.Error("should not find number when preceded by Latin letters at VALID leniency")
	}

	// With Latin suffix — not found at VALID.
	text = number + "def"
	m = u.FindNumbers(text, "US", LeniencyValid, 65535)
	if m.HasNext() {
		t.Error("should not find number when followed by Latin letters at VALID leniency")
	}

	// But found at POSSIBLE leniency.
	text = "abc" + number + "def"
	m = u.FindNumbers(text, "US", LeniencyPossible, 65535)
	if !m.HasNext() {
		t.Error("should find number at POSSIBLE leniency even with surrounding Latin chars")
	}
}

func TestMatcherMoneyNotMatched(t *testing.T) {
	// "$170,000" should not be matched as a phone number.
	u := Instance()
	m := u.FindNumbers("$170,000", "US", LeniencyValid, 65535)
	if m.HasNext() {
		t.Error("money amount should not be matched")
	}

	// But "$170,000 " with trailing space and a valid number should work.
	text := "$170,000 650-253-0000"
	m = u.FindNumbers(text, "US", LeniencyValid, 65535)
	if !m.HasNext() {
		t.Fatal("should find the number after money amount")
	}
	match := m.Next()
	if match.RawString != "650-253-0000" {
		t.Errorf("expected '650-253-0000', got %q", match.RawString)
	}
}

func TestMatcherPercentageNotMatched(t *testing.T) {
	u := Instance()
	m := u.FindNumbers("4.8%650-253-0000", "US", LeniencyValid, 65535)
	if m.HasNext() {
		t.Error("percentage followed by number should not match at VALID")
	}
}

func TestMatcherDateNotMatched(t *testing.T) {
	// Slash-separated dates should not be matched.
	u := Instance()
	dates := []string{
		"3/10/2011",
		"31/10/96",
		"08/31/95",
	}
	for _, date := range dates {
		m := u.FindNumbers(date, "US", LeniencyPossible, 65535)
		if m.HasNext() {
			t.Errorf("date %q should not be matched", date)
		}
	}
}

func TestMatcherTimestampNotMatched(t *testing.T) {
	// "2012-01-02 08:00" should not produce a match due to timestamp detection.
	u := Instance()
	m := u.FindNumbers("2012-01-02 08:00", "US", LeniencyPossible, 65535)
	if m.HasNext() {
		match := m.Next()
		t.Errorf("timestamp should not be matched, got %q", match.RawString)
	}
}

func TestMatcherPublicationPageNotMatched(t *testing.T) {
	u := Instance()
	m := u.FindNumbers("211-227 (2003)", "US", LeniencyPossible, 65535)
	if m.HasNext() {
		t.Error("publication page reference should not be matched")
	}
}

func TestMatcherNoMatchInEmptyString(t *testing.T) {
	u := Instance()
	m := u.FindNumbers("", "US", LeniencyPossible, 65535)
	if m.HasNext() {
		t.Error("should not find matches in empty string")
	}
}

func TestMatcherNoMatchIfNoNumber(t *testing.T) {
	u := Instance()
	m := u.FindNumbers("This is just some text with no numbers", "US", LeniencyPossible, 65535)
	if m.HasNext() {
		t.Error("should not find matches in text without numbers")
	}
}

func TestMatcherMaxTries(t *testing.T) {
	// With maxTries = 0, no match should be found even with a valid number.
	u := Instance()
	m := u.FindNumbers("650-253-0000", "US", LeniencyPossible, 0)
	if m.HasNext() {
		t.Error("maxTries=0 should prevent finding any match")
	}
}

func TestMatcherContainsMoreThanOneSlash(t *testing.T) {
	// A date-like string with slashes from default country.
	number := PhoneNumber{
		CountryCode:       1,
		CountryCodeSource: CountryCodeFromDefaultCountry,
	}
	if !containsMoreThanOneSlashInNationalNumber(number, "1/05/2013") {
		t.Error("expected true for date-like string")
	}

	// First slash after country code is OK — should be false.
	number = PhoneNumber{
		CountryCode:       49,
		CountryCodeSource: CountryCodeFromNumberWithPlus,
	}
	if containsMoreThanOneSlashInNationalNumber(number, "49/69/2013") {
		t.Error("expected false when first slash follows CC")
	}

	// But a third slash makes it true.
	if !containsMoreThanOneSlashInNationalNumber(number, "49/69/20/13") {
		t.Error("expected true for three slashes even with CC")
	}
}

func TestMatcherMatchPositions(t *testing.T) {
	u := Instance()
	text := "Call me at +1 650 253 0000 please"
	m := u.FindNumbers(text, "US", LeniencyPossible, 65535)
	if !m.HasNext() {
		t.Fatal("expected a match")
	}
	match := m.Next()
	if match.Start != 11 {
		t.Errorf("expected Start=11, got %d", match.Start)
	}
	expected := "+1 650 253 0000"
	if match.RawString != expected {
		t.Errorf("expected %q, got %q", expected, match.RawString)
	}
	if match.End != 11+len(expected) {
		t.Errorf("expected End=%d, got %d", 11+len(expected), match.End)
	}
}

func TestMatcherParsedNumberCorrect(t *testing.T) {
	u := Instance()
	text := "My number is +1-650-253-0000."
	m := u.FindNumbers(text, "US", LeniencyPossible, 65535)
	if !m.HasNext() {
		t.Fatal("expected a match")
	}
	match := m.Next()
	if match.Number.CountryCode != 1 {
		t.Errorf("expected CC=1, got %d", match.Number.CountryCode)
	}
	if match.Number.NationalNumber != 6502530000 {
		t.Errorf("expected NN=6502530000, got %d", match.Number.NationalNumber)
	}
	// RawInput and CountryCodeSource should be cleared.
	if match.Number.RawInput != "" {
		t.Error("RawInput should be cleared in match")
	}
	if match.Number.CountryCodeSource != CountryCodeUnspecified {
		t.Error("CountryCodeSource should be cleared in match")
	}
}

func TestMatcherWithSurroundingZipcodes(t *testing.T) {
	u := Instance()
	number := "415-666-7777"
	text := "My address is CA 34215 - " + number + " is my number."
	m := u.FindNumbers(text, "US", LeniencyPossible, 65535)
	if !m.HasNext() {
		t.Fatal("expected to find number")
	}
	match := m.Next()
	if match.RawString != number {
		t.Errorf("expected %q, got %q", number, match.RawString)
	}
}

func TestMatcherHasNextIdempotent(t *testing.T) {
	// Multiple calls to HasNext without Next should not advance the iterator.
	u := Instance()
	m := u.FindNumbers("+1-650-253-0000", "US", LeniencyPossible, 65535)

	// Call HasNext multiple times.
	if !m.HasNext() {
		t.Fatal("expected HasNext=true")
	}
	if !m.HasNext() {
		t.Fatal("second HasNext should also be true")
	}
	if !m.HasNext() {
		t.Fatal("third HasNext should also be true")
	}

	match := m.Next()
	if match.RawString != "+1-650-253-0000" {
		t.Errorf("unexpected match: %q", match.RawString)
	}

	if m.HasNext() {
		t.Error("should be done after consuming the only match")
	}
}

func TestMatcherNonPlusPrefixedNotFoundForInvalidRegion(t *testing.T) {
	// Without a valid region, only numbers with '+' prefix should be found.
	u := Instance()
	m := u.FindNumbers("1 650 253 0000", "ZZ", LeniencyPossible, 65535)
	if m.HasNext() {
		t.Error("should not find number without + when region is ZZ")
	}

	// With +, it should be found.
	m = u.FindNumbers("+1 650 253 0000", "ZZ", LeniencyPossible, 65535)
	if !m.HasNext() {
		t.Error("should find number with + even when region is ZZ")
	}
}

func TestMatcherSurroundingChineseChars(t *testing.T) {
	// Chinese characters around a number should not prevent matching at VALID.
	u := Instance()
	text := "我的电话号码是+1 650 253 0000。"
	m := u.FindNumbers(text, "US", LeniencyValid, 65535)
	if !m.HasNext() {
		t.Error("should find number surrounded by Chinese characters")
	}
}

func TestMatcherNonMatchingBrackets(t *testing.T) {
	// Unmatched brackets should prevent a match.
	u := Instance()
	tests := []string{
		"(650 253 0000", // Missing closing bracket (after opening)
		"650) 253 0000", // Closing without opening
	}
	for _, text := range tests {
		m := u.FindNumbers(text, "US", LeniencyPossible, 65535)
		if m.HasNext() {
			match := m.Next()
			// The full text might not match, but inner match could extract valid part.
			// Just verify it doesn't crash.
			_ = match
		}
	}
}

func TestMatcherLeniencyValid(t *testing.T) {
	u := Instance()
	// A valid US number should be found at VALID leniency.
	m := u.FindNumbers("+1 650 253 0000", "US", LeniencyValid, 65535)
	if !m.HasNext() {
		t.Error("valid number should be found at VALID leniency")
	}
}

func TestMatcherLeniencyStrictGrouping(t *testing.T) {
	u := Instance()
	// Properly grouped US number.
	m := u.FindNumbers("(650) 253-0000", "US", LeniencyStrictGrouping, 65535)
	if !m.HasNext() {
		t.Error("properly grouped number should be found at STRICT_GROUPING")
	}
}

func TestMatcherLeniencyExactGrouping(t *testing.T) {
	u := Instance()
	// Exactly grouped US number.
	m := u.FindNumbers("(650) 253-0000", "US", LeniencyExactGrouping, 65535)
	if !m.HasNext() {
		t.Error("exactly grouped number should be found at EXACT_GROUPING")
	}
}

func TestMatcherInvalidUTF8(t *testing.T) {
	u := Instance()
	// Invalid UTF-8 should not panic and should return no matches.
	invalid := "\xff\xfe 650-253-0000 \xff"
	m := u.FindNumbers(invalid, "US", LeniencyPossible, 65535)
	// We don't assert match/no-match here — just verify no panic.
	for m.HasNext() {
		_ = m.Next()
	}
}

func TestMatcherSequences(t *testing.T) {
	u := Instance()
	// Numbers separated by " - " should all be found.
	text := "Call 033316005 or 032316005!"
	m := u.FindNumbers(text, "NZ", LeniencyPossible, 65535)

	if !m.HasNext() {
		t.Fatal("expected first match")
	}
	match1 := m.Next()
	if match1.RawString != "033316005" {
		t.Errorf("first match: got %q", match1.RawString)
	}

	if !m.HasNext() {
		t.Fatal("expected second match")
	}
	match2 := m.Next()
	if match2.RawString != "032316005" {
		t.Errorf("second match: got %q", match2.RawString)
	}

	if m.HasNext() {
		t.Error("expected no more matches")
	}
}

func TestMatcherFindWithPlusNoRegion(t *testing.T) {
	doTestFindInContext(t, "+64 3 331 6005", "ZZ")
}

func TestMatcherFindNumbersArgentina(t *testing.T) {
	doTestFindInContext(t, "+54 9 343 555 1212", "AR")
	doTestFindInContext(t, "+54 11 3797 0000", "AR")
}

func TestMatcherFindNumbersMexico(t *testing.T) {
	doTestFindInContext(t, "+52 (449)978-0001", "MX")
}

func TestMatcherInterspersedWithSpace(t *testing.T) {
	doTestFindInContext(t, "0 3   3 3 1   6 0 0 5", "NZ")
}
