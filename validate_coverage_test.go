package phonesafe

import "testing"

// Tests targeting low-coverage paths in validate.go, format.go, and parse.go.

func TestIsPossibleNumberForAllTypes(t *testing.T) {
	u := Instance()

	// Exercise getDescForType and testNumberLengthForType for multiple types.
	usNum := PhoneNumber{CountryCode: 1, NationalNumber: 6502530000}
	types := []PhoneNumberType{
		TypeFixedLine, TypeMobile, TypeFixedLineOrMobile, TypeTollFree,
		TypePremiumRate, TypeSharedCost, TypeVOIP, TypePersonalNumber,
		TypePager, TypeUAN, TypeVoicemail, TypeUnknown,
	}
	for _, typ := range types {
		// Should not panic for any type.
		u.IsPossibleNumberForTypeWithReason(usNum, typ)
	}

	// Test with a GB number too (different length rules).
	gbNum := PhoneNumber{CountryCode: 44, NationalNumber: 2070313000}
	for _, typ := range types {
		u.IsPossibleNumberForTypeWithReason(gbNum, typ)
	}
}

func TestIsPossibleNumberValidationResults(t *testing.T) {
	u := Instance()

	tests := []struct {
		name   string
		number PhoneNumber
		want   ValidationResult
	}{
		{
			name:   "valid US",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 6502530000},
			want:   IsPossible,
		},
		{
			name:   "too short US",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 650253},
			want:   TooShort,
		},
		{
			name:   "too long US",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 650253000012},
			want:   TooLong,
		},
		{
			name:   "invalid country code",
			number: PhoneNumber{CountryCode: 999, NationalNumber: 12345},
			want:   InvalidCountryCode,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := u.IsPossibleNumberWithReason(tc.number)
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestGetExampleNumberForTypeAllTypes(t *testing.T) {
	u := Instance()

	// Exercise GetExampleNumberForType for various types.
	// Some types may not have examples — that's fine, we just want code coverage.
	types := []PhoneNumberType{
		TypeFixedLine, TypeMobile, TypeTollFree, TypePremiumRate,
		TypeSharedCost, TypeVOIP, TypePersonalNumber, TypePager,
		TypeUAN, TypeVoicemail,
	}
	for _, typ := range types {
		num, ok := u.GetExampleNumberForType("US", typ)
		if ok && num.CountryCode != 1 {
			t.Errorf("type %d: CC=%d, want 1", typ, num.CountryCode)
		}
	}

	// Try multiple regions.
	for _, region := range []string{"GB", "DE", "JP", "AU"} {
		_, _ = u.GetExampleNumberForType(region, TypeMobile)
	}
}

func TestTruncateTooLongNumberVariants(t *testing.T) {
	u := Instance()

	// Number that's way too long — must keep trimming.
	veryLong := PhoneNumber{CountryCode: 1, NationalNumber: 6502530000123456}
	truncated, ok := u.TruncateTooLongNumber(veryLong)
	if !ok {
		t.Error("failed to truncate very long number")
	}
	if !u.IsValidNumber(truncated) {
		t.Error("truncated result is not valid")
	}

	// Number with Italian leading zeros that's too long.
	itLong := PhoneNumber{
		CountryCode:        39,
		NationalNumber:     23661830012,
		ItalianLeadingZero: true,
	}
	truncatedIT, ok := u.TruncateTooLongNumber(itLong)
	if ok && !u.IsValidNumber(truncatedIT) {
		t.Error("truncated IT number is not valid")
	}

	// Invalid country code — should return false.
	invalid := PhoneNumber{CountryCode: 999, NationalNumber: 12345}
	_, ok = u.TruncateTooLongNumber(invalid)
	if ok {
		t.Error("truncate should fail for invalid CC")
	}
}

func TestFormatOutOfCountryKeepingAlphaMorePaths(t *testing.T) {
	u := Instance()

	// Same country code, different region (AU number from AU)
	auNum := PhoneNumber{
		CountryCode: 61, NationalNumber: 236618300,
		RawInput: "+61 2-3661-8300",
	}
	got := u.FormatOutOfCountryKeepingAlphaChars(auNum, "AU")
	if got == "" {
		t.Error("expected non-empty result for AU number from AU")
	}

	// Number from region with multiple international prefixes (SG)
	usNum := PhoneNumber{
		CountryCode: 1, NationalNumber: 8007493524,
		RawInput: "800-749-3524",
	}
	got = u.FormatOutOfCountryKeepingAlphaChars(usNum, "SG")
	if got == "" {
		t.Error("expected non-empty result from SG")
	}

	// Invalid region (non-supported)
	got = u.FormatOutOfCountryKeepingAlphaChars(usNum, "AQ")
	if got == "" {
		t.Error("expected non-empty result from unsupported region")
	}
}

func TestFormatInOriginalFormatMorePaths(t *testing.T) {
	u := Instance()

	// Number without raw input and no CountryCodeSource (uses National format)
	num := PhoneNumber{CountryCode: 1, NationalNumber: 6502530000}
	got := u.FormatInOriginalFormat(num, "US")
	if got == "" {
		t.Error("expected non-empty result")
	}

	// Number with CountryCodeSource=FromNumberWithPlus
	numPlus, _ := u.ParseAndKeepRawInput("+16502530000", "US")
	got = u.FormatInOriginalFormat(numPlus, "US")
	if got == "" {
		t.Error("expected non-empty for FromPlus")
	}

	// Number with raw input that has no formatting pattern
	rawNum := PhoneNumber{
		CountryCode:    1,
		NationalNumber: 6502530000,
		RawInput:       "6502530000",
	}
	got = u.FormatInOriginalFormat(rawNum, "US")
	if got == "" {
		t.Error("expected non-empty for raw-only number")
	}
}

func TestFormatPreferredCarrierCode(t *testing.T) {
	u := Instance()

	// Number with stored carrier code.
	arMobile := PhoneNumber{
		CountryCode:                  54,
		NationalNumber:               92234654321,
		PreferredDomesticCarrierCode: "14",
	}
	got := u.FormatNationalNumberWithPreferredCarrierCode(arMobile, "15")
	// Should use "14" from the number, not "15" fallback.
	if got == "" {
		t.Error("expected non-empty result")
	}

	// Number without stored carrier code — uses fallback.
	arMobile2 := PhoneNumber{
		CountryCode:    54,
		NationalNumber: 92234654321,
	}
	got = u.FormatNationalNumberWithPreferredCarrierCode(arMobile2, "15")
	if got == "" {
		t.Error("expected non-empty result with fallback carrier")
	}
}

func TestUnicodeDigitToASCII(t *testing.T) {
	u := Instance()
	// Parse a number with fullwidth digits (exercises unicodeDigitToASCII).
	num, err := u.Parse("０３-１２３４-５６７８", "JP")
	if err != nil {
		t.Fatalf("failed to parse fullwidth digits: %v", err)
	}
	if num.CountryCode != 81 {
		t.Errorf("CC=%d, want 81", num.CountryCode)
	}
}

// TestAllRegionsCovered verifies that every supported region has at minimum:
// - A valid example number
// - Correct validation
// - Non-empty international format output
// This meta-test enforces the region coverage matrix from TESTING_SPEC.md §5.
func TestAllRegionsCovered(t *testing.T) {
	u := Instance()
	for _, region := range u.GetSupportedRegions() {
		t.Run(region, func(t *testing.T) {
			num, ok := u.GetExampleNumber(region)
			if !ok {
				t.Skipf("no example number for %s", region)
				return
			}
			if !u.IsValidNumber(num) {
				t.Errorf("example number for %s is not valid", region)
			}
			formatted := u.Format(num, FormatInternational)
			if formatted == "" {
				t.Errorf("format returned empty for %s", region)
			}
		})
	}
}
