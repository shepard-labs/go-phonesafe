package phonesafe

import "testing"

// Test numbers matching upstream test fixtures.
var (
	nzNumber = PhoneNumber{CountryCode: 64, NationalNumber: 33316005}
	usNumber = PhoneNumber{CountryCode: 1, NationalNumber: 6502530000}
)

func TestIsNumberMatchExact(t *testing.T) {
	u := Instance()
	tests := []struct {
		name   string
		first  string
		second string
		want   MatchType
	}{
		{"formatting differs", "+64 3 331 6005", "+64 03 331 6005", MatchExact},
		{"intl no spaces", "+800 1234 5678", "+80012345678", MatchExact},
		{"dashes vs spaces", "+64 03 331-6005", "+64 03331 6005", MatchExact},
		{"cc compacted", "+643 331-6005", "+64033316005", MatchExact},
		{"no leading zero", "+643 331-6005", "+6433316005", MatchExact},
		{"spaces vs compact", "+64 3 331-6005", "+6433316005", MatchExact},
		{"RFC3966 isub ignored", "+64 3 331-6005", "tel:+64-3-331-6005;isub=123", MatchExact},
		{"alpha number", "+1800 siX-Flags", "+1 800 7493 5247", MatchExact},
		{"ext extn vs #", "+64 3 331-6005 extn 1234", "+6433316005#1234", MatchExact},
		{"ext ext. vs ;", "+64 3 331-6005 ext. 1234", "+6433316005;1234", MatchExact},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := u.IsNumberMatchString(tt.first, tt.second)
			if got != tt.want {
				t.Errorf("IsNumberMatchString(%q, %q) = %d, want %d", tt.first, tt.second, got, tt.want)
			}
		})
	}
}

func TestIsNumberMatchProto(t *testing.T) {
	u := Instance()

	// Parsed number vs string.
	got := u.isNumberMatchOneString(nzNumber, "+6403 331 6005")
	if got != MatchExact {
		t.Errorf("IsNumberMatch(NZ_NUMBER, %q) = %d, want MatchExact", "+6403 331 6005", got)
	}

	// Number with extension.
	nzExt := nzNumber
	nzExt.Extension = "3456"
	got = u.isNumberMatchOneString(nzExt, "+643 331 6005 ext 3456")
	if got != MatchExact {
		t.Errorf("IsNumberMatch(nz+ext, string+ext) = %d, want MatchExact", got)
	}

	// Empty extension is ignored.
	nzEmpty := nzNumber
	nzEmpty.Extension = ""
	got = u.isNumberMatchOneString(nzEmpty, "+6403 331 6005")
	if got != MatchExact {
		t.Errorf("IsNumberMatch(nz_empty_ext, string) = %d, want MatchExact", got)
	}

	// Two protos, one with empty extension.
	got = u.IsNumberMatch(nzEmpty, nzNumber)
	if got != MatchExact {
		t.Errorf("IsNumberMatch(nz_empty_ext, nz) = %d, want MatchExact", got)
	}
}

func TestIsNumberMatchShortNSNDiffLeadingZeros(t *testing.T) {
	u := Instance()

	one := PhoneNumber{CountryCode: 64, NationalNumber: 33316005, ItalianLeadingZero: true}
	two := PhoneNumber{CountryCode: 64, NationalNumber: 33316005, ItalianLeadingZero: true, NumberOfLeadingZeros: 2}
	got := u.IsNumberMatch(one, two)
	if got != MatchShortNSN {
		t.Errorf("diff leading zeros = %d, want MatchShortNSN", got)
	}

	// When ItalianLeadingZero is false for one, leading zeros count is irrelevant for matching NSN.
	one2 := PhoneNumber{CountryCode: 64, NationalNumber: 33316005, ItalianLeadingZero: false, NumberOfLeadingZeros: 1}
	two2 := PhoneNumber{CountryCode: 64, NationalNumber: 33316005, ItalianLeadingZero: true, NumberOfLeadingZeros: 1}
	got = u.IsNumberMatch(one2, two2)
	if got != MatchShortNSN {
		t.Errorf("one without ItalianLeadingZero = %d, want MatchShortNSN", got)
	}
}

func TestIsNumberMatchAcceptsProtoDefaultsAsMatch(t *testing.T) {
	u := Instance()

	one := PhoneNumber{CountryCode: 64, NationalNumber: 33316005, ItalianLeadingZero: true}
	// NumberOfLeadingZeros = 1 is the default when ItalianLeadingZero is true.
	two := PhoneNumber{CountryCode: 64, NationalNumber: 33316005, ItalianLeadingZero: true, NumberOfLeadingZeros: 1}
	got := u.IsNumberMatch(one, two)
	if got != MatchExact {
		t.Errorf("default leading zeros = %d, want MatchExact", got)
	}
}

func TestIsNumberMatchDiffLeadingZerosIfItalianLeadingZeroFalse(t *testing.T) {
	u := Instance()

	one := PhoneNumber{CountryCode: 64, NationalNumber: 33316005}
	two := PhoneNumber{CountryCode: 64, NationalNumber: 33316005, NumberOfLeadingZeros: 1}
	got := u.IsNumberMatch(one, two)
	if got != MatchExact {
		t.Errorf("leading zeros ignored when ItalianLeadingZero false = %d, want MatchExact", got)
	}

	// Even if set to 10, still equivalent since ItalianLeadingZero is false for both.
	two.NumberOfLeadingZeros = 10
	got = u.IsNumberMatch(one, two)
	if got != MatchExact {
		t.Errorf("leading zeros=10 ignored when ItalianLeadingZero false = %d, want MatchExact", got)
	}
}

func TestIsNumberMatchIgnoresSomeFields(t *testing.T) {
	u := Instance()

	// RawInput, CountryCodeSource, and PreferredDomesticCarrierCode should be ignored.
	br1 := PhoneNumber{
		CountryCode:                  55,
		NationalNumber:               3121286979,
		CountryCodeSource:            CountryCodeFromNumberWithPlus,
		PreferredDomesticCarrierCode: "12",
		RawInput:                     "012 3121286979",
	}
	br2 := PhoneNumber{
		CountryCode:                  55,
		NationalNumber:               3121286979,
		CountryCodeSource:            CountryCodeFromDefaultCountry,
		PreferredDomesticCarrierCode: "14",
		RawInput:                     "143121286979",
	}
	got := u.IsNumberMatch(br1, br2)
	if got != MatchExact {
		t.Errorf("ignored fields differ = %d, want MatchExact", got)
	}
}

func TestIsNumberMatchNonMatches(t *testing.T) {
	u := Instance()
	tests := []struct {
		name   string
		first  string
		second string
		want   MatchType
	}{
		{"different last digit", "03 331 6005", "03 331 6006", MatchNone},
		{"different CC", "+800 1234 5678", "+1 800 1234 5678", MatchNone},
		{"CC 64 vs 1+64", "+64 3 331-6005", "+16433316005", MatchNone},
		{"CC 64 vs 61", "+64 3 331-6005", "+6133316005", MatchNone},
		{"different ext", "+64 3 331-6005 extn 1234", "0116433316005#1235", MatchNone},
		{"different ext rfc3966", "+64 3 331-6005 extn 1234", "tel:+64-3-331-6005;ext=1235", MatchNone},
		{"nsn match but ext differs", "+64 3 331-6005 ext.1235", "3 331 6005#1234", MatchNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := u.IsNumberMatchString(tt.first, tt.second)
			if got != tt.want {
				t.Errorf("IsNumberMatchString(%q, %q) = %d, want %d", tt.first, tt.second, got, tt.want)
			}
		})
	}
}

func TestIsNumberMatchInvalidNumbers(t *testing.T) {
	u := Instance()
	tests := []struct {
		name   string
		first  string
		second string
	}{
		{"too short", "4", "3 331 6043"},
		{"just CC", "+43", "+64 3 331 6005"},
		{"just CC no plus", "+43", "64 3 331 6005"},
		{"not a number", "Dog", "64 3 331 6005"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := u.IsNumberMatchString(tt.first, tt.second)
			if got != MatchInvalidNumber {
				t.Errorf("IsNumberMatchString(%q, %q) = %d, want MatchInvalidNumber", tt.first, tt.second, got)
			}
		})
	}
}

func TestIsNumberMatchNSNMatches(t *testing.T) {
	u := Instance()
	tests := []struct {
		name   string
		first  string
		second string
		want   MatchType
	}{
		{"intl vs national", "+64 3 331-6005", "03 331 6005", MatchNSN},
		{"intl vs rfc3966 phone-context", "+64 3 331-6005", "tel:03-331-6005;isub=1234;phone-context=abc.nz", MatchNSN},
		{"US intl vs national with 1", "+1 650-253 0000", "1 650 253 0000", MatchNSN},
		{"US intl vs no prefix", "+1 650-253 0000", "6502530000", MatchNSN},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := u.IsNumberMatchString(tt.first, tt.second)
			if got != tt.want {
				t.Errorf("IsNumberMatchString(%q, %q) = %d, want %d", tt.first, tt.second, got, tt.want)
			}
		})
	}

	// Proto vs string.
	got := u.isNumberMatchOneString(nzNumber, "03 331 6005")
	if got != MatchNSN {
		t.Errorf("NZ proto vs national string = %d, want MatchNSN", got)
	}

	// US proto vs string with leading 1.
	got = u.isNumberMatchOneString(usNumber, "1-650-253-0000")
	if got != MatchNSN {
		t.Errorf("US proto vs '1-650-253-0000' = %d, want MatchNSN", got)
	}

	got = u.isNumberMatchOneString(usNumber, "6502530000")
	if got != MatchNSN {
		t.Errorf("US proto vs '6502530000' = %d, want MatchNSN", got)
	}
}

func TestIsNumberMatchShortNSNMatches(t *testing.T) {
	u := Instance()
	tests := []struct {
		name   string
		first  string
		second string
		want   MatchType
	}{
		{"intl vs partial", "+64 3 331-6005", "331 6005", MatchShortNSN},
		{"intl vs rfc3966 partial", "+64 3 331-6005", "tel:331-6005;phone-context=abc.nz", MatchShortNSN},
		{"intl vs rfc3966 isub", "+64 3 331-6005", "tel:331-6005;isub=1234;phone-context=abc.nz", MatchShortNSN},
		{"intl vs rfc3966 params", "+64 3 331-6005", "tel:331-6005;isub=1234;phone-context=abc.nz;a=%A1", MatchShortNSN},
		{"national prefix unknown", "3 331-6005", "03 331 6005", MatchShortNSN},
		{"partial vs partial", "3 331-6005", "331 6005", MatchShortNSN},
		{"partial vs rfc3966", "3 331-6005", "tel:331-6005;phone-context=abc.nz", MatchShortNSN},
		{"partial vs intl", "3 331-6005", "+64 331 6005", MatchShortNSN},
		{"national vs partial", "03 331-6005", "331 6005", MatchShortNSN},
		{"US suffix match", "1 234 345 6789", "345 6789", MatchShortNSN},
		{"US parens suffix", "+1 (234) 345 6789", "345 6789", MatchShortNSN},
		{"ext present one side", "+64 3 331-6005", "3 331 6005#1234", MatchShortNSN},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := u.IsNumberMatchString(tt.first, tt.second)
			if got != tt.want {
				t.Errorf("IsNumberMatchString(%q, %q) = %d, want %d", tt.first, tt.second, got, tt.want)
			}
		})
	}
}

func TestIsNumberMatchItalianLeadingZero(t *testing.T) {
	u := Instance()

	// One has Italian leading zero, one does not.
	it1 := PhoneNumber{CountryCode: 39, NationalNumber: 1234, ItalianLeadingZero: true}
	it2 := PhoneNumber{CountryCode: 39, NationalNumber: 1234}
	got := u.IsNumberMatch(it1, it2)
	if got != MatchShortNSN {
		t.Errorf("Italian leading zero diff = %d, want MatchShortNSN", got)
	}
}

func TestIsNumberMatchExtensionHandling(t *testing.T) {
	u := Instance()

	// One has extension, other has empty extension — should be SHORT_NSN_MATCH.
	it1 := PhoneNumber{CountryCode: 39, NationalNumber: 1234, Extension: "1234"}
	it2 := PhoneNumber{CountryCode: 39, NationalNumber: 1234, Extension: ""}
	got := u.IsNumberMatch(it1, it2)
	if got != MatchShortNSN {
		t.Errorf("extension vs empty = %d, want MatchShortNSN", got)
	}
}

func TestIsNumberMatchShortNSNProtoVsString(t *testing.T) {
	u := Instance()

	// For this case the match is SHORT_NSN because the country codes differ and
	// we cannot assume the "1" is a national prefix.
	randomNumber := PhoneNumber{CountryCode: 41, NationalNumber: 6502530000}
	got := u.isNumberMatchOneString(randomNumber, "1-650-253-0000")
	if got != MatchShortNSN {
		t.Errorf("CH number vs US-formatted string = %d, want MatchShortNSN", got)
	}
}
