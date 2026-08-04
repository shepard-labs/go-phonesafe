package phonesafe

import "testing"

// Test numbers for coverage tests (prefixed cov_ to avoid conflicts).
var (
	covUSNum  = PhoneNumber{CountryCode: 1, NationalNumber: 6502530000}
	covUSLong = PhoneNumber{CountryCode: 1, NationalNumber: 65025300001}
	covGBNum  = PhoneNumber{CountryCode: 44, NationalNumber: 2070313000}
	covDENum  = PhoneNumber{CountryCode: 49, NationalNumber: 30123456}
	covITNum  = PhoneNumber{CountryCode: 39, NationalNumber: 236618300, ItalianLeadingZero: true}
	covBSNum  = PhoneNumber{CountryCode: 1, NationalNumber: 2423651234}
	covMXNum1 = PhoneNumber{CountryCode: 52, NationalNumber: 3312345678}
)

func TestFormatByPattern(t *testing.T) {
	u := Instance()

	tests := []struct {
		name   string
		number PhoneNumber
		format PhoneNumberFormat
		rules  []NumberFormatRule
		want   string
	}{
		{
			name:   "US national",
			number: covUSNum,
			format: FormatNational,
			rules: []NumberFormatRule{{
				Pattern: `(\d{3})(\d{3})(\d{4})`,
				Format:  "($1) $2-$3",
			}},
			want: "(650) 253-0000",
		},
		{
			name:   "US international",
			number: covUSNum,
			format: FormatInternational,
			rules: []NumberFormatRule{{
				Pattern: `(\d{3})(\d{3})(\d{4})`,
				Format:  "($1) $2-$3",
			}},
			want: "+1 (650) 253-0000",
		},
		{
			name:   "US RFC3966",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 6507129823},
			format: FormatRFC3966,
			rules: []NumberFormatRule{{
				Pattern: `(\d{3})(\d{3})(\d{4})`,
				Format:  "($1) $2-$3",
			}},
			want: "tel:+1-650-712-9823",
		},
		{
			name:   "IT national with pattern",
			number: covITNum,
			format: FormatNational,
			rules: []NumberFormatRule{{
				Pattern: `(\d{2})(\d{5})(\d{3})`,
				Format:  "$1-$2 $3",
			}},
			want: "02-36618 300",
		},
		{
			name:   "IT international with pattern",
			number: covITNum,
			format: FormatInternational,
			rules: []NumberFormatRule{{
				Pattern: `(\d{2})(\d{5})(\d{3})`,
				Format:  "$1-$2 $3",
			}},
			want: "+39 02-36618 300",
		},
		{
			name:   "GB with NP rule $NP$FG",
			number: covGBNum,
			format: FormatNational,
			rules: []NumberFormatRule{{
				Pattern:                      `(\d{2})(\d{4})(\d{4})`,
				Format:                       "$1 $2 $3",
				NationalPrefixFormattingRule: "$NP$FG",
			}},
			want: "020 7031 3000",
		},
		{
			name:   "GB with NP rule ($NP$FG)",
			number: covGBNum,
			format: FormatNational,
			rules: []NumberFormatRule{{
				Pattern:                      `(\d{2})(\d{4})(\d{4})`,
				Format:                       "$1 $2 $3",
				NationalPrefixFormattingRule: "($NP$FG)",
			}},
			want: "(020) 7031 3000",
		},
		{
			name:   "GB no NP rule",
			number: covGBNum,
			format: FormatNational,
			rules: []NumberFormatRule{{
				Pattern: `(\d{2})(\d{4})(\d{4})`,
				Format:  "$1 $2 $3",
			}},
			want: "20 7031 3000",
		},
		{
			name:   "GB international no NP rule",
			number: covGBNum,
			format: FormatInternational,
			rules: []NumberFormatRule{{
				Pattern: `(\d{2})(\d{4})(\d{4})`,
				Format:  "$1 $2 $3",
			}},
			want: "+44 20 7031 3000",
		},
		{
			name:   "BS with NP rule",
			number: covBSNum,
			format: FormatNational,
			rules: []NumberFormatRule{{
				Pattern:                      `(\d{3})(\d{3})(\d{4})`,
				Format:                       "$1 $2-$3",
				NationalPrefixFormattingRule: "$NP ($FG)",
			}},
			want: "1 (242) 365-1234",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := u.FormatByPattern(tc.number, tc.format, tc.rules)
			if got != tc.want {
				t.Errorf("FormatByPattern = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatOutOfCountryKeepingAlphaChars(t *testing.T) {
	u := Instance()

	tests := []struct {
		name        string
		number      PhoneNumber
		callingFrom string
		want        string
	}{
		{
			name: "US number with raw from AU",
			number: PhoneNumber{
				CountryCode: 1, NationalNumber: 8007493524,
				RawInput: "1800 749-3524",
			},
			callingFrom: "AU",
			want:        "0011 1 800 749-3524",
		},
		{
			name: "US number with raw dashes from AU",
			number: PhoneNumber{
				CountryCode: 1, NationalNumber: 8007493524,
				RawInput: "1-800-749-3524",
			},
			callingFrom: "AU",
			want:        "0011 1 800-749-3524",
		},
		{
			name: "US number from US (NANPA)",
			number: PhoneNumber{
				CountryCode: 1, NationalNumber: 8007493524,
				RawInput: "800 749-3524",
			},
			callingFrom: "US",
			want:        "1 800 749-3524",
		},
		{
			name: "no raw input falls back to formatOutOfCountry",
			number: PhoneNumber{
				CountryCode: 1, NationalNumber: 8007493524,
			},
			callingFrom: "DE",
			want:        "00 1 800-749-3524",
		},
		{
			name: "invalid CC uses raw input",
			number: PhoneNumber{
				CountryCode: 0, NationalNumber: 18007493524,
				RawInput: "1-800-SIX-flag",
			},
			callingFrom: "DE",
			want:        "1-800-SIX-flag",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := u.FormatOutOfCountryKeepingAlphaChars(tc.number, tc.callingFrom)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatInOriginalFormatExtended(t *testing.T) {
	u := Instance()

	tests := []struct {
		name   string
		input  string
		region string
		from   string
		want   string
	}{
		{
			name:   "GB international with +",
			input:  "+442087654321",
			region: "GB",
			from:   "GB",
			want:   "+44 20 8765 4321",
		},
		{
			name:   "GB national with trunk",
			input:  "02087654321",
			region: "GB",
			from:   "GB",
			want:   "020 8765 4321",
		},
		{
			name:   "GB from US via IDD",
			input:  "011442087654321",
			region: "US",
			from:   "US",
			want:   "011 44 20 8765 4321",
		},
		{
			name:   "GB without plus",
			input:  "442087654321",
			region: "GB",
			from:   "GB",
			want:   "44 20 8765 4321",
		},
		{
			name:   "US with national prefix",
			input:  "18003456789",
			region: "US",
			from:   "US",
			want:   "1 800-345-6789",
		},
		{
			name:   "GB without national prefix",
			input:  "2087654321",
			region: "GB",
			from:   "GB",
			want:   "20 8765 4321",
		},
		{
			name:   "IT fixed line",
			input:  "0212345678",
			region: "IT",
			from:   "IT",
			want:   "02 1234 5678",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			num, err := u.ParseAndKeepRawInput(tc.input, tc.region)
			if err != nil {
				t.Fatalf("ParseAndKeepRawInput(%q, %q) error: %v", tc.input, tc.region, err)
			}
			got := u.FormatInOriginalFormat(num, tc.from)
			if got != tc.want {
				t.Errorf("FormatInOriginalFormat(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestFormatNumberForMobileDialingExtended(t *testing.T) {
	u := Instance()

	tests := []struct {
		name    string
		number  PhoneNumber
		from    string
		withFmt bool
		want    string
	}{
		{
			name:   "DE national from DE no format",
			number: covDENum,
			from:   "DE",
			want:   "030123456",
		},
		{
			name:   "DE from CH no format",
			number: covDENum,
			from:   "CH",
			want:   "+4930123456",
		},
		{
			name:    "US with formatting from US",
			number:  usNumber,
			from:    "US",
			withFmt: true,
			want:    "+1 650-253-0000",
		},
		{
			name:   "US no format from US",
			number: covUSNum,
			from:   "US",
			want:   "+16502530000",
		},
		{
			name:   "US no format from CA (NANPA cross-region)",
			number: covUSNum,
			from:   "CA",
			want:   "+16502530000",
		},
		{
			name:   "US no format from BR (international)",
			number: covUSNum,
			from:   "BR",
			want:   "+16502530000",
		},
		{
			name:   "MX fixed from MX",
			number: covMXNum1,
			from:   "MX",
			want:   "+523312345678",
		},
		{
			name:   "MX fixed from US",
			number: covMXNum1,
			from:   "US",
			want:   "+523312345678",
		},
		{
			name:   "US long number",
			number: covUSLong,
			from:   "US",
			want:   "+165025300001",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := u.FormatNumberForMobileDialing(tc.number, tc.from, tc.withFmt)
			if got != tc.want {
				t.Errorf("FormatNumberForMobileDialing = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNormalizeHelper(t *testing.T) {
	// Exercise normalizeHelper via FormatOutOfCountryKeepingAlphaChars
	u := Instance()
	num := PhoneNumber{
		CountryCode: 1, NationalNumber: 8007493524,
		RawInput: "1800 749-3524",
	}
	got := u.FormatOutOfCountryKeepingAlphaChars(num, "AU")
	if got != "0011 1 800 749-3524" {
		t.Errorf("got %q, want %q", got, "0011 1 800 749-3524")
	}
}

func TestIsAlphaNumber(t *testing.T) {
	u := Instance()
	tests := []struct {
		input string
		want  bool
	}{
		{"1-800-SIX-FLAGS", true},
		{"1-800-123-4567", false},
		{"ab", false},
		{"+1 800 SIX-FLAG", true},
		{"", false},
	}
	for _, tc := range tests {
		got := u.IsAlphaNumber(tc.input)
		if got != tc.want {
			t.Errorf("IsAlphaNumber(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestGetRegionCodesForCountryCode(t *testing.T) {
	u := Instance()
	// NANPA has multiple regions
	regions := u.GetRegionCodesForCountryCode(1)
	if len(regions) < 2 {
		t.Fatalf("expected multiple regions for CC=1, got %d", len(regions))
	}
	if regions[0] != "US" {
		t.Errorf("first region for CC=1 = %q, want US", regions[0])
	}

	// Unknown CC
	if got := u.GetRegionCodesForCountryCode(999); got != nil {
		t.Errorf("CC=999 should return nil, got %v", got)
	}
}

func TestParseErrorMessage(t *testing.T) {
	u := Instance()
	_, err := u.Parse("not-a-number", "US")
	if err == nil {
		t.Fatal("expected error")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected *ParseError, got %T", err)
	}
	msg := pe.Error()
	if msg == "" {
		t.Error("ParseError.Error() returned empty string")
	}
}

func TestDescHasPossibleNumberData(t *testing.T) {
	// Exercise via IsPossibleNumberForType with a type that has [-1] lengths
	u := Instance()
	// Pager type in US should have no data ([-1])
	num := PhoneNumber{CountryCode: 1, NationalNumber: 6502530000}
	_ = u.IsPossibleNumberForTypeWithReason(num, TypePager)
}
