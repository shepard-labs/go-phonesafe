package phonesafe

import (
	"slices"
	"sort"
	"testing"
)

func TestGetRegionCodeForCountryCode(t *testing.T) {
	u := Instance()
	tests := []struct {
		cc   int
		want string
	}{
		{1, "US"},
		{44, "GB"},
		{49, "DE"},
		{81, "JP"},
		{999, ""},
	}
	for _, tc := range tests {
		got := u.GetRegionCodeForCountryCode(tc.cc)
		if got != tc.want {
			t.Errorf("GetRegionCodeForCountryCode(%d) = %q, want %q", tc.cc, got, tc.want)
		}
	}
}

func TestGetSupportedRegions(t *testing.T) {
	u := Instance()
	regions := u.GetSupportedRegions()

	if len(regions) < 200 {
		t.Fatalf("expected 200+ regions, got %d", len(regions))
	}
	if !sort.StringsAreSorted(regions) {
		t.Error("GetSupportedRegions() returned unsorted slice")
	}
	// Must not contain "001" (non-geo entity).
	if slices.Contains(regions, "001") {
		t.Error("GetSupportedRegions() should not contain '001'")
	}
	// Must contain well-known regions.
	for _, r := range []string{"US", "GB", "DE", "JP", "BR"} {
		if !slices.Contains(regions, r) {
			t.Errorf("GetSupportedRegions() missing %q", r)
		}
	}
}

func TestGetSupportedCallingCodes(t *testing.T) {
	u := Instance()
	codes := u.GetSupportedCallingCodes()

	if len(codes) < 200 {
		t.Fatalf("expected 200+ calling codes, got %d", len(codes))
	}
	if !sort.IntsAreSorted(codes) {
		t.Error("GetSupportedCallingCodes() returned unsorted slice")
	}
	for _, cc := range []int{1, 44, 49, 81, 86} {
		if !slices.Contains(codes, cc) {
			t.Errorf("GetSupportedCallingCodes() missing %d", cc)
		}
	}
}

func TestGetSupportedTypesForRegion(t *testing.T) {
	u := Instance()

	// Valid region should return types.
	usTypes := u.GetSupportedTypesForRegion("US")
	if len(usTypes) == 0 {
		t.Fatal("GetSupportedTypesForRegion('US') returned empty")
	}
	hasFixed := slices.Contains(usTypes, TypeFixedLine)
	hasMobile := slices.Contains(usTypes, TypeMobile)
	hasTollFree := slices.Contains(usTypes, TypeTollFree)
	if !hasFixed || !hasMobile || !hasTollFree {
		t.Errorf("US types missing expected: fixed=%v mobile=%v tollFree=%v",
			hasFixed, hasMobile, hasTollFree)
	}

	// Invalid region returns nil.
	if got := u.GetSupportedTypesForRegion("ZZ"); got != nil {
		t.Errorf("GetSupportedTypesForRegion('ZZ') = %v, want nil", got)
	}
}

func TestIsMobileNumberPortableRegion(t *testing.T) {
	u := Instance()
	// US supports MNP.
	if !u.IsMobileNumberPortableRegion("US") {
		t.Error("IsMobileNumberPortableRegion('US') should be true")
	}
	// Invalid region.
	if u.IsMobileNumberPortableRegion("ZZ") {
		t.Error("IsMobileNumberPortableRegion('ZZ') should be false")
	}
}

func TestGetExampleNumber(t *testing.T) {
	u := Instance()

	num, ok := u.GetExampleNumber("US")
	if !ok {
		t.Fatal("GetExampleNumber('US') returned false")
	}
	if num.CountryCode != 1 {
		t.Errorf("example US number CC=%d, want 1", num.CountryCode)
	}
	if !u.IsValidNumber(num) {
		t.Error("GetExampleNumber('US') returned invalid number")
	}

	// Invalid region.
	_, ok = u.GetExampleNumber("ZZ")
	if ok {
		t.Error("GetExampleNumber('ZZ') should return false")
	}
}

func TestGetExampleNumberForType(t *testing.T) {
	u := Instance()

	num, ok := u.GetExampleNumberForType("US", TypeMobile)
	if !ok {
		t.Fatal("GetExampleNumberForType('US', TypeMobile) returned false")
	}
	if num.CountryCode != 1 {
		t.Errorf("example US mobile CC=%d, want 1", num.CountryCode)
	}
	if !u.IsValidNumber(num) {
		t.Error("GetExampleNumberForType('US', TypeMobile) returned invalid number")
	}
	// Verify it's actually a mobile number (or fixed-line-or-mobile for US).
	numType := u.GetNumberType(num)
	if numType != TypeMobile && numType != TypeFixedLineOrMobile {
		t.Errorf("example US mobile type=%d, want TypeMobile or TypeFixedLineOrMobile", numType)
	}
}

func TestGetInvalidExampleNumber(t *testing.T) {
	u := Instance()

	num, ok := u.GetInvalidExampleNumber("US")
	if !ok {
		t.Fatal("GetInvalidExampleNumber('US') returned false")
	}
	if num.CountryCode != 1 {
		t.Errorf("invalid example US CC=%d, want 1", num.CountryCode)
	}
	if u.IsValidNumber(num) {
		t.Error("GetInvalidExampleNumber('US') returned a valid number")
	}
}

func TestTruncateTooLongNumber(t *testing.T) {
	u := Instance()

	// Already valid number — should return same.
	valid, _ := u.Parse("+1 650 253 0000", "US")
	got, ok := u.TruncateTooLongNumber(valid)
	if !ok {
		t.Fatal("TruncateTooLongNumber on valid number returned false")
	}
	if got.NationalNumber != valid.NationalNumber {
		t.Error("TruncateTooLongNumber changed a valid number")
	}

	// Too-long US number (11 digits instead of 10).
	tooLong := PhoneNumber{CountryCode: 1, NationalNumber: 65025300001}
	truncated, ok := u.TruncateTooLongNumber(tooLong)
	if !ok {
		t.Fatal("TruncateTooLongNumber failed to truncate")
	}
	if !u.IsValidNumber(truncated) {
		t.Error("truncated number is not valid")
	}
	if truncated.NationalNumber >= tooLong.NationalNumber {
		t.Error("truncated number is not shorter")
	}
}

func TestGetLengthOfGeographicalAreaCode(t *testing.T) {
	u := Instance()

	// US number: area code is 3 digits.
	usNum, _ := u.Parse("+1 650 253 0000", "US")
	if got := u.GetLengthOfGeographicalAreaCode(usNum); got != 3 {
		t.Errorf("area code length for US number = %d, want 3", got)
	}

	// Non-geographical number should return 0.
	tollFree, _ := u.Parse("+1 800 234 5678", "US")
	numType := u.GetNumberType(tollFree)
	if numType == TypeTollFree {
		// Toll-free is not geographical.
		if got := u.GetLengthOfGeographicalAreaCode(tollFree); got != 0 {
			t.Errorf("area code length for toll-free = %d, want 0", got)
		}
	}
}

func TestGetLengthOfNationalDestinationCode(t *testing.T) {
	u := Instance()

	// US number: NDC is 3 digits (area code).
	usNum, _ := u.Parse("+1 650 253 0000", "US")
	if got := u.GetLengthOfNationalDestinationCode(usNum); got != 3 {
		t.Errorf("NDC length for US number = %d, want 3", got)
	}

	// UK London number: NDC is 2 digits ("20").
	gbNum, _ := u.Parse("+44 20 7031 3000", "GB")
	if got := u.GetLengthOfNationalDestinationCode(gbNum); got != 2 {
		t.Errorf("NDC length for GB London = %d, want 2", got)
	}
}
