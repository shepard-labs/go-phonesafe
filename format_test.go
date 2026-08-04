package phonesafe

import (
	"os"
	"strings"
	"testing"
)

// Test number constants matching upstream.
var (
	fmtUsNumber    = PhoneNumber{CountryCode: 1, NationalNumber: 6502530000}
	fmtUsTollFree  = PhoneNumber{CountryCode: 1, NationalNumber: 8002530000}
	fmtUsPremium   = PhoneNumber{CountryCode: 1, NationalNumber: 9002530000}
	fmtUsSpoof     = PhoneNumber{CountryCode: 1, NationalNumber: 0}
	fmtUsSpoofRaw  = PhoneNumber{CountryCode: 1, NationalNumber: 0, RawInput: "000-000-0000"}
	fmtBsNumber    = PhoneNumber{CountryCode: 1, NationalNumber: 2423651234}
	fmtGbNumber    = PhoneNumber{CountryCode: 44, NationalNumber: 2070313000}
	fmtGbMobile    = PhoneNumber{CountryCode: 44, NationalNumber: 7912345678}
	fmtDeNumber    = PhoneNumber{CountryCode: 49, NationalNumber: 30123456}
	fmtDeShort     = PhoneNumber{CountryCode: 49, NationalNumber: 1234}
	fmtItNumber    = PhoneNumber{CountryCode: 39, NationalNumber: 236618300, ItalianLeadingZero: true}
	fmtItMobile    = PhoneNumber{CountryCode: 39, NationalNumber: 345678901}
	fmtAuNumber    = PhoneNumber{CountryCode: 61, NationalNumber: 236618300}
	fmtArNumber    = PhoneNumber{CountryCode: 54, NationalNumber: 1187654321}
	fmtArMobile    = PhoneNumber{CountryCode: 54, NationalNumber: 91187654321}
	fmtNzNumber    = PhoneNumber{CountryCode: 64, NationalNumber: 33316005}
	fmtSgNumber    = PhoneNumber{CountryCode: 65, NationalNumber: 65218000}
	fmtIntlTolFree = PhoneNumber{CountryCode: 800, NationalNumber: 12345678}
	fmtUnknownCC   = PhoneNumber{CountryCode: 2, NationalNumber: 12345}
)

func TestFormatE164(t *testing.T) {
	u := Instance()

	tests := []struct {
		number   PhoneNumber
		expected string
	}{
		{fmtUsNumber, "+16502530000"},
		{fmtDeNumber, "+4930123456"},
		{fmtItNumber, "+390236618300"},
		{fmtItMobile, "+39345678901"},
		{fmtAuNumber, "+61236618300"},
		{fmtArNumber, "+541187654321"},
		{fmtArMobile, "+5491187654321"},
		{fmtNzNumber, "+6433316005"},
		{fmtIntlTolFree, "+80012345678"},
	}

	for _, tt := range tests {
		got := u.Format(tt.number, FormatE164)
		if got != tt.expected {
			t.Errorf("Format(%+v, E164) = %q, want %q", tt.number, got, tt.expected)
		}
	}
}

func TestFormatNational(t *testing.T) {
	u := Instance()

	tests := []struct {
		name     string
		number   PhoneNumber
		expected string
	}{
		{"US", fmtUsNumber, "(650) 253-0000"},
		{"US TollFree", fmtUsTollFree, "(800) 253-0000"},
		{"US Premium", fmtUsPremium, "(900) 253-0000"},
		{"US Spoof with raw", fmtUsSpoofRaw, "000-000-0000"},
		{"US Spoof no raw", fmtUsSpoof, "0"},
		{"GB fixed", fmtGbNumber, "020 7031 3000"},
		{"GB mobile", fmtGbMobile, "07912 345678"},
		{"DE", fmtDeNumber, "030 123456"},
		{"DE short", fmtDeShort, "1234"},
		{"IT fixed", fmtItNumber, "02 3661 8300"},
		{"IT mobile", fmtItMobile, "345 678 901"},
		{"AU", fmtAuNumber, "(02) 3661 8300"},
		{"AR fixed", fmtArNumber, "011 8765-4321"},
		{"NZ", fmtNzNumber, "03 331 6005"},
		{"SG", fmtSgNumber, "6521 8000"},
		{"BS", fmtBsNumber, "(242) 365-1234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := u.Format(tt.number, FormatNational)
			if got != tt.expected {
				t.Errorf("Format National = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFormatInternational(t *testing.T) {
	u := Instance()

	tests := []struct {
		name     string
		number   PhoneNumber
		expected string
	}{
		{"US", fmtUsNumber, "+1 650-253-0000"},
		{"GB fixed", fmtGbNumber, "+44 20 7031 3000"},
		{"GB mobile", fmtGbMobile, "+44 7912 345678"},
		{"DE", fmtDeNumber, "+49 30 123456"},
		{"DE short", fmtDeShort, "+49 1234"},
		{"IT fixed", fmtItNumber, "+39 02 3661 8300"},
		{"IT mobile", fmtItMobile, "+39 345 678 901"},
		{"AU", fmtAuNumber, "+61 2 3661 8300"},
		{"AR fixed", fmtArNumber, "+54 11 8765-4321"},
		{"AR mobile", fmtArMobile, "+54 9 11 8765-4321"},
		{"NZ", fmtNzNumber, "+64 3 331 6005"},
		{"SG", fmtSgNumber, "+65 6521 8000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := u.Format(tt.number, FormatInternational)
			if got != tt.expected {
				t.Errorf("Format International = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFormatRFC3966(t *testing.T) {
	u := Instance()

	tests := []struct {
		name     string
		number   PhoneNumber
		expected string
	}{
		{"US", fmtUsPremium, "tel:+1-900-253-0000"},
		{"DE", fmtDeNumber, "tel:+49-30-123456"},
		{"NZ", fmtNzNumber, "tel:+64-3-331-6005"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := u.Format(tt.number, FormatRFC3966)
			if got != tt.expected {
				t.Errorf("Format RFC3966 = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFormatWithExtension(t *testing.T) {
	u := Instance()

	// NZ with extension (default prefix " ext. ")
	nzWithExt := fmtNzNumber
	nzWithExt.Extension = "1234"
	if got := u.Format(nzWithExt, FormatNational); got != "03 331 6005 ext. 1234" {
		t.Errorf("NZ national with ext = %q, want %q", got, "03 331 6005 ext. 1234")
	}
	if got := u.Format(nzWithExt, FormatRFC3966); got != "tel:+64-3-331-6005;ext=1234" {
		t.Errorf("NZ RFC3966 with ext = %q, want %q", got, "tel:+64-3-331-6005;ext=1234")
	}

	// GB with extension (preferred prefix " x")
	gbWithExt := fmtGbNumber
	gbWithExt.Extension = "4567"
	if got := u.Format(gbWithExt, FormatNational); got != "020 7031 3000 x4567" {
		t.Errorf("GB national with ext = %q, want %q", got, "020 7031 3000 x4567")
	}
}

func TestFormatInvalidCountryCode(t *testing.T) {
	u := Instance()
	// Number with invalid country code — should just return the NSN.
	got := u.Format(fmtUnknownCC, FormatNational)
	if got != "12345" {
		t.Errorf("Unknown CC national = %q, want %q", got, "12345")
	}
}

func TestFormatOutOfCountryCallingNumber(t *testing.T) {
	u := Instance()

	tests := []struct {
		name        string
		number      PhoneNumber
		callingFrom string
		expected    string
	}{
		{"US from DE", fmtUsPremium, "DE", "00 1 900-253-0000"},
		{"US from BS (NANPA)", fmtUsNumber, "BS", "1 (650) 253-0000"},
		{"US from PL", fmtUsNumber, "PL", "00 1 650-253-0000"},
		{"GB mobile from US", fmtGbMobile, "US", "011 44 7912 345678"},
		{"DE short from GB", fmtDeShort, "GB", "00 49 1234"},
		{"DE short from DE (same CC)", fmtDeShort, "DE", "1234"},
		{"IT from US", fmtItNumber, "US", "011 39 02 3661 8300"},
		{"IT from IT (same CC)", fmtItNumber, "IT", "02 3661 8300"},
		{"SG from SG (same CC)", fmtSgNumber, "SG", "6521 8000"},
		{"AR mobile from US", fmtArMobile, "US", "011 54 9 11 8765-4321"},
		{"Intl toll-free from US", fmtIntlTolFree, "US", "011 800 1234 5678"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := u.FormatOutOfCountryCallingNumber(tt.number, tt.callingFrom)
			if got != tt.expected {
				t.Errorf("FormatOutOfCountry(%q) = %q, want %q", tt.callingFrom, got, tt.expected)
			}
		})
	}
}

func TestFormatOutOfCountryWithInvalidRegion(t *testing.T) {
	u := Instance()

	// Invalid region falls back to international format.
	got := u.FormatOutOfCountryCallingNumber(fmtUsNumber, "AQ")
	if got != "+1 650-253-0000" {
		t.Errorf("US from AQ = %q, want %q", got, "+1 650-253-0000")
	}
}

func TestFormatOutOfCountryWithPreferredIntlPrefix(t *testing.T) {
	u := Instance()

	// AU has preferred international prefix "0011".
	if got := u.FormatOutOfCountryCallingNumber(fmtItNumber, "AU"); got != "0011 39 02 3661 8300" {
		t.Errorf("IT from AU = %q, want %q", got, "0011 39 02 3661 8300")
	}
}

func TestFormatNationalNumberWithCarrierCode(t *testing.T) {
	u := Instance()

	// US: carrier code has no effect (no DomesticCarrierCodeFormattingRule in US).
	got := u.FormatNationalNumberWithCarrierCode(fmtUsNumber, "15")
	if got != "(650) 253-0000" {
		t.Errorf("US with carrier = %q, want %q", got, "(650) 253-0000")
	}
}

func TestFormatNumberForMobileDialing(t *testing.T) {
	u := Instance()

	// From within NANPA, use international if dialable.
	got := u.FormatNumberForMobileDialing(fmtUsNumber, "US", true)
	expected := "+1 650-253-0000"
	if got != expected {
		t.Errorf("US from US with formatting = %q, want %q", got, expected)
	}

	// Without formatting.
	got = u.FormatNumberForMobileDialing(fmtUsNumber, "US", false)
	if got != "+16502530000" {
		t.Errorf("US from US without formatting = %q, want %q", got, "+16502530000")
	}

	// International call with formatting.
	got = u.FormatNumberForMobileDialing(fmtUsNumber, "GB", true)
	if got != "+1 650-253-0000" {
		t.Errorf("US from GB with formatting = %q, want %q", got, "+1 650-253-0000")
	}

	// International call without formatting.
	got = u.FormatNumberForMobileDialing(fmtUsNumber, "GB", false)
	if got != "+16502530000" {
		t.Errorf("US from GB without formatting = %q, want %q", got, "+16502530000")
	}
}

func TestFormatNumberForMobileDialingBR(t *testing.T) {
	u := Instance()

	// Brazil mobile: 11 9XXXX-XXXX (valid pattern).
	brMobile := PhoneNumber{CountryCode: 55, NationalNumber: 11961234567}

	// Without carrier code — should return empty for BR mobile/fixed.
	got := u.FormatNumberForMobileDialing(brMobile, "BR", true)
	if got != "" {
		t.Errorf("BR without carrier from BR = %q, want empty", got)
	}

	// With preferred carrier code — should return non-empty.
	brWithCarrier := PhoneNumber{
		CountryCode:                  55,
		NationalNumber:               11961234567,
		PreferredDomesticCarrierCode: "12",
	}
	got = u.FormatNumberForMobileDialing(brWithCarrier, "BR", true)
	if got == "" {
		t.Error("BR with carrier from BR should not be empty")
	}
}

func TestFormatInOriginalFormat(t *testing.T) {
	u := Instance()

	// Number with + prefix should format internationally.
	num1, err := u.ParseAndKeepRawInput("+442087654321", "GB")
	if err != nil {
		t.Fatal(err)
	}
	got := u.FormatInOriginalFormat(num1, "GB")
	if got != "+44 20 8765 4321" {
		t.Errorf("FormatInOriginalFormat(+44...) = %q, want %q", got, "+44 20 8765 4321")
	}

	// Number with national prefix should format nationally.
	num2, err := u.ParseAndKeepRawInput("02087654321", "GB")
	if err != nil {
		t.Fatal(err)
	}
	got = u.FormatInOriginalFormat(num2, "GB")
	if got != "020 8765 4321" {
		t.Errorf("FormatInOriginalFormat(020...) = %q, want %q", got, "020 8765 4321")
	}

	// US number in E164.
	num3, err := u.ParseAndKeepRawInput("+16502530000", "US")
	if err != nil {
		t.Fatal(err)
	}
	got = u.FormatInOriginalFormat(num3, "US")
	if got != "+1 650-253-0000" {
		t.Errorf("FormatInOriginalFormat(+1...) = %q, want %q", got, "+1 650-253-0000")
	}
}

func TestFormatRoundTrip(t *testing.T) {
	u := Instance()

	// Parse(Format(number, E164), "ZZ") should give back the same number.
	numbers := []PhoneNumber{
		fmtUsNumber, fmtGbNumber, fmtGbMobile, fmtDeNumber, fmtItNumber,
		fmtItMobile, fmtAuNumber, fmtArNumber, fmtArMobile, fmtNzNumber, fmtSgNumber,
	}

	for _, num := range numbers {
		e164 := u.Format(num, FormatE164)
		parsed, err := u.Parse(e164, "ZZ")
		if err != nil {
			t.Errorf("Parse(%q, ZZ) error: %v", e164, err)
			continue
		}
		if parsed.CountryCode != num.CountryCode || parsed.NationalNumber != num.NationalNumber {
			t.Errorf("Round-trip failed for %+v: got %+v via %q", num, parsed, e164)
		}
	}
}

func TestIsNANPACountry(t *testing.T) {
	u := Instance()

	if !u.IsNANPACountry("US") {
		t.Error("US should be NANPA")
	}
	if !u.IsNANPACountry("CA") {
		t.Error("CA should be NANPA")
	}
	if !u.IsNANPACountry("BS") {
		t.Error("BS should be NANPA")
	}
	if u.IsNANPACountry("GB") {
		t.Error("GB should not be NANPA")
	}
	if u.IsNANPACountry("DE") {
		t.Error("DE should not be NANPA")
	}
}

func TestGetNddPrefixForRegion(t *testing.T) {
	u := Instance()

	if got := u.GetNddPrefixForRegion("US", false); got != "1" {
		t.Errorf("US NDD = %q, want %q", got, "1")
	}
	if got := u.GetNddPrefixForRegion("GB", false); got != "0" {
		t.Errorf("GB NDD = %q, want %q", got, "0")
	}
	if got := u.GetNddPrefixForRegion("ZZ", false); got != "" {
		t.Errorf("ZZ NDD = %q, want empty", got)
	}
}

func TestGetCountryMobileToken(t *testing.T) {
	if got := GetCountryMobileToken(54); got != "9" {
		t.Errorf("AR mobile token = %q, want %q", got, "9")
	}
	if got := GetCountryMobileToken(1); got != "" {
		t.Errorf("US mobile token = %q, want empty", got)
	}
}

// --- Golden File Tests ---
// These tests verify formatting output against golden files in testdata/format_golden/.
// See TESTING_SPEC.md §3.3 for the golden file test pattern.

func TestFormatGoldenFiles(t *testing.T) {
	u := Instance()

	t.Run("us_national", func(t *testing.T) {
		testFormatGoldenFile(t, u, "testdata/format_golden/us_national.txt", FormatNational)
	})
	t.Run("international", func(t *testing.T) {
		testFormatGoldenFile(t, u, "testdata/format_golden/international.txt", FormatInternational)
	})
}

func testFormatGoldenFile(t *testing.T, u *PhoneNumberUtil, path string, format PhoneNumberFormat) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", path, err)
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			t.Fatalf("%s:%d: invalid format, expected input|region|expected", path, i+1)
		}
		input, region, expected := parts[0], parts[1], parts[2]

		num, parseErr := u.Parse(input, region)
		if parseErr != nil {
			t.Errorf("%s:%d: Parse(%q, %q) error: %v", path, i+1, input, region, parseErr)
			continue
		}
		got := u.Format(num, format)
		if got != expected {
			t.Errorf("%s:%d: Format(%q, %v) = %q, want %q", path, i+1, input, format, got, expected)
		}
	}
}

func TestFormatGoldenAsYouType(t *testing.T) {
	u := Instance()

	data, err := os.ReadFile("testdata/format_golden/as_you_type_us.txt")
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	var f *AsYouTypeFormatter
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if line == "CLEAR" {
			f = u.NewAsYouTypeFormatter("US")
			continue
		}
		if f == nil {
			f = u.NewAsYouTypeFormatter("US")
		}

		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			t.Fatalf("as_you_type_us.txt:%d: invalid format, expected digit|expected", i+1)
		}
		digitStr, expected := parts[0], parts[1]
		// \s at end of expected = trailing space (can't be represented literally in text files).
		expected = strings.ReplaceAll(expected, `\s`, " ")
		if len(digitStr) != 1 {
			t.Fatalf("as_you_type_us.txt:%d: digit must be single char, got %q", i+1, digitStr)
		}
		digit := rune(digitStr[0])
		got := f.InputDigit(digit)
		if got != expected {
			t.Errorf("as_you_type_us.txt:%d: InputDigit(%q) = %q, want %q",
				i+1, string(digit), got, expected)
		}
	}
}
