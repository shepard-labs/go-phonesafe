package phonesafe

import "testing"

// --- IsValidNumber ---

func TestIsValidNumber(t *testing.T) {
	u := Instance()

	tests := []struct {
		name   string
		number PhoneNumber
		want   bool
	}{
		{
			name:   "US valid",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 6502530000},
			want:   true,
		},
		{
			name:   "IT valid with leading zero",
			number: PhoneNumber{CountryCode: 39, NationalNumber: 236618300, ItalianLeadingZero: true},
			want:   true,
		},
		{
			name:   "GB mobile valid",
			number: PhoneNumber{CountryCode: 44, NationalNumber: 7912345678},
			want:   true,
		},
		{
			name:   "NZ mobile valid",
			number: PhoneNumber{CountryCode: 64, NationalNumber: 21387835},
			want:   true,
		},
		{
			name:   "International toll-free valid",
			number: PhoneNumber{CountryCode: 800, NationalNumber: 12345678},
			want:   true,
		},
		{
			name:   "Universal premium rate valid",
			number: PhoneNumber{CountryCode: 979, NationalNumber: 123456789},
			want:   true,
		},
		// Invalid numbers
		{
			name:   "US too short",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 2530000},
			want:   false,
		},
		{
			name:   "IT too long",
			number: PhoneNumber{CountryCode: 39, NationalNumber: 23661830000, ItalianLeadingZero: true},
			want:   false,
		},
		{
			name:   "GB mobile too short",
			number: PhoneNumber{CountryCode: 44, NationalNumber: 791234567},
			want:   false,
		},
		{
			name:   "DE too short",
			number: PhoneNumber{CountryCode: 49, NationalNumber: 1234},
			want:   false,
		},
		{
			name:   "NZ invalid",
			number: PhoneNumber{CountryCode: 64, NationalNumber: 3316005},
			want:   false,
		},
		{
			name:   "Invalid country code 3923",
			number: PhoneNumber{CountryCode: 3923, NationalNumber: 2366},
			want:   false,
		},
		{
			name:   "Country code zero",
			number: PhoneNumber{CountryCode: 0, NationalNumber: 2366},
			want:   false,
		},
		{
			name:   "Intl toll-free too long",
			number: PhoneNumber{CountryCode: 800, NationalNumber: 123456789},
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := u.IsValidNumber(tc.number)
			if got != tc.want {
				t.Errorf("IsValidNumber(%+v) = %v, want %v", tc.number, got, tc.want)
			}
		})
	}
}

// --- IsValidNumberForRegion ---

func TestIsValidNumberForRegion(t *testing.T) {
	u := Instance()

	tests := []struct {
		name   string
		number PhoneNumber
		region string
		want   bool
	}{
		{
			name:   "Bahamas number valid for BS",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 2423232345},
			region: "BS",
			want:   true,
		},
		{
			name:   "Bahamas number invalid for US",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 2423232345},
			region: "US",
			want:   false,
		},
		{
			name:   "Reunion number valid for RE",
			number: PhoneNumber{CountryCode: 262, NationalNumber: 262123456},
			region: "RE",
			want:   true,
		},
		{
			name:   "Reunion number invalid for YT",
			number: PhoneNumber{CountryCode: 262, NationalNumber: 262123456},
			region: "YT",
			want:   false,
		},
		{
			name:   "Mayotte number valid for YT",
			number: PhoneNumber{CountryCode: 262, NationalNumber: 269601234},
			region: "YT",
			want:   true,
		},
		{
			name:   "Mayotte number invalid for RE",
			number: PhoneNumber{CountryCode: 262, NationalNumber: 269601234},
			region: "RE",
			want:   false,
		},
		{
			name:   "Intl toll-free valid for 001",
			number: PhoneNumber{CountryCode: 800, NationalNumber: 12345678},
			region: "001",
			want:   true,
		},
		{
			name:   "Intl toll-free invalid for US",
			number: PhoneNumber{CountryCode: 800, NationalNumber: 12345678},
			region: "US",
			want:   false,
		},
		{
			name:   "Invalid CC for ZZ",
			number: PhoneNumber{CountryCode: 3923, NationalNumber: 2366},
			region: "ZZ",
			want:   false,
		},
		{
			name:   "Invalid CC for 001",
			number: PhoneNumber{CountryCode: 3923, NationalNumber: 2366},
			region: "001",
			want:   false,
		},
		{
			name:   "Zero CC for 001",
			number: PhoneNumber{CountryCode: 0, NationalNumber: 2366},
			region: "001",
			want:   false,
		},
		{
			name:   "Zero CC for ZZ",
			number: PhoneNumber{CountryCode: 0, NationalNumber: 2366},
			region: "ZZ",
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := u.IsValidNumberForRegion(tc.number, tc.region)
			if got != tc.want {
				t.Errorf("IsValidNumberForRegion(%+v, %q) = %v, want %v",
					tc.number, tc.region, got, tc.want)
			}
		})
	}
}

// --- IsPossibleNumber ---

func TestIsPossibleNumber(t *testing.T) {
	u := Instance()

	tests := []struct {
		name   string
		number PhoneNumber
		want   bool
	}{
		{
			name:   "US 10-digit",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 6502530000},
			want:   true,
		},
		{
			name:   "US 7-digit (local only)",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 2530000},
			want:   true, // IsPossibleLocalOnly counts as possible
		},
		{
			name:   "GB landline",
			number: PhoneNumber{CountryCode: 44, NationalNumber: 2070313000},
			want:   true,
		},
		{
			name:   "Intl toll-free",
			number: PhoneNumber{CountryCode: 800, NationalNumber: 12345678},
			want:   true,
		},
		// Not possible
		{
			name:   "US too short",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 253000},
			want:   false,
		},
		{
			name:   "US too long",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 65025300000},
			want:   false,
		},
		{
			name:   "Invalid CC",
			number: PhoneNumber{CountryCode: 0, NationalNumber: 2530000},
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := u.IsPossibleNumber(tc.number)
			if got != tc.want {
				t.Errorf("IsPossibleNumber(%+v) = %v, want %v", tc.number, got, tc.want)
			}
		})
	}
}

// --- IsPossibleNumberWithReason ---

func TestIsPossibleNumberWithReason(t *testing.T) {
	u := Instance()

	tests := []struct {
		name   string
		number PhoneNumber
		want   ValidationResult
	}{
		{
			name:   "US valid 10-digit",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 6502530000},
			want:   IsPossible,
		},
		{
			name:   "US 7-digit local only",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 2530000},
			want:   IsPossibleLocalOnly,
		},
		{
			name:   "CC zero = invalid country code",
			number: PhoneNumber{CountryCode: 0, NationalNumber: 2530000},
			want:   InvalidCountryCode,
		},
		{
			name:   "US too short",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 253000},
			want:   TooShort,
		},
		{
			name:   "US too long",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 65025300000},
			want:   TooLong,
		},
		{
			name:   "GB valid",
			number: PhoneNumber{CountryCode: 44, NationalNumber: 2070310000},
			want:   IsPossible,
		},
		{
			name:   "DE valid 8-digit",
			number: PhoneNumber{CountryCode: 49, NationalNumber: 30123456},
			want:   IsPossible,
		},
		{
			name:   "SG 10-digit",
			number: PhoneNumber{CountryCode: 65, NationalNumber: 1234567890},
			want:   IsPossible,
		},
		{
			name:   "Intl toll-free too long (9 digits)",
			number: PhoneNumber{CountryCode: 800, NationalNumber: 123456789},
			want:   TooLong,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := u.IsPossibleNumberWithReason(tc.number)
			if got != tc.want {
				t.Errorf("IsPossibleNumberWithReason(%+v) = %v, want %v",
					tc.number, got, tc.want)
			}
		})
	}
}

// --- IsPossibleNumberForTypeWithReason ---

func TestIsPossibleNumberForTypeWithReason(t *testing.T) {
	u := Instance()

	tests := []struct {
		name    string
		number  PhoneNumber
		numType PhoneNumberType
		want    ValidationResult
	}{
		// Argentina — different type lengths
		// AR FixedLine: national=[10], local=[6,7,8]
		// AR Mobile: national=[10,11], local=[6,7,8]
		// AR TollFree: national=[10,11]
		{
			name:    "AR 5-digit too short for unknown",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 12345},
			numType: TypeUnknown,
			want:    TooShort,
		},
		{
			name:    "AR 5-digit too short for fixed-line",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 12345},
			numType: TypeFixedLine,
			want:    TooShort,
		},
		{
			name:    "AR 6-digit local only for unknown",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 123456},
			numType: TypeUnknown,
			want:    IsPossibleLocalOnly,
		},
		{
			name:    "AR 6-digit local only for fixed-line",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 123456},
			numType: TypeFixedLine,
			want:    IsPossibleLocalOnly,
		},
		{
			name:    "AR 6-digit local only for mobile",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 123456},
			numType: TypeMobile,
			want:    IsPossibleLocalOnly,
		},
		{
			name:    "AR 6-digit too short for toll-free",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 123456},
			numType: TypeTollFree,
			want:    TooShort,
		},
		{
			name:    "AR 10-digit possible for unknown",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 1234567890},
			numType: TypeUnknown,
			want:    IsPossible,
		},
		{
			name:    "AR 10-digit possible for fixed-line",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 1234567890},
			numType: TypeFixedLine,
			want:    IsPossible,
		},
		{
			name:    "AR 10-digit possible for mobile",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 1234567890},
			numType: TypeMobile,
			want:    IsPossible,
		},
		{
			name:    "AR 10-digit possible for toll-free",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 1234567890},
			numType: TypeTollFree,
			want:    IsPossible,
		},
		{
			name:    "AR 11-digit possible for unknown",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 12345678901},
			numType: TypeUnknown,
			want:    IsPossible,
		},
		{
			name:    "AR 11-digit too long for fixed-line",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 12345678901},
			numType: TypeFixedLine,
			want:    TooLong,
		},
		{
			name:    "AR 11-digit possible for mobile",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 12345678901},
			numType: TypeMobile,
			want:    IsPossible,
		},
		{
			name:    "AR 11-digit possible for toll-free",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 12345678901},
			numType: TypeTollFree,
			want:    IsPossible,
		},
		// Germany — local-only lengths
		{
			name:    "DE 2-digit local only for unknown",
			number:  PhoneNumber{CountryCode: 49, NationalNumber: 12},
			numType: TypeUnknown,
			want:    IsPossibleLocalOnly,
		},
		{
			name:    "DE 2-digit local only for fixed-line",
			number:  PhoneNumber{CountryCode: 49, NationalNumber: 12},
			numType: TypeFixedLine,
			want:    IsPossibleLocalOnly,
		},
		{
			name:    "DE 2-digit too short for mobile",
			number:  PhoneNumber{CountryCode: 49, NationalNumber: 12},
			numType: TypeMobile,
			want:    TooShort,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := u.IsPossibleNumberForTypeWithReason(tc.number, tc.numType)
			if got != tc.want {
				t.Errorf("IsPossibleNumberForTypeWithReason(%+v, %d) = %v, want %v",
					tc.number, tc.numType, got, tc.want)
			}
		})
	}
}

// --- IsPossibleNumberForType ---

func TestIsPossibleNumberForType(t *testing.T) {
	u := Instance()

	tests := []struct {
		name    string
		number  PhoneNumber
		numType PhoneNumberType
		want    bool
	}{
		{
			name:    "AR 6-digit possible for fixed-line (local only)",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 123456},
			numType: TypeFixedLine,
			want:    true, // IsPossibleLocalOnly counts as possible
		},
		{
			name:    "AR 6-digit possible for mobile (local only)",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 123456},
			numType: TypeMobile,
			want:    true, // IsPossibleLocalOnly counts as possible
		},
		{
			name:    "AR 6-digit not possible for toll-free",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 123456},
			numType: TypeTollFree,
			want:    false,
		},
		{
			name:    "AR 10-digit possible for mobile",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 1234567890},
			numType: TypeMobile,
			want:    true,
		},
		{
			name:    "AR 11-digit possible for mobile",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 12345678901},
			numType: TypeMobile,
			want:    true,
		},
		{
			name:    "AR 11-digit not possible for fixed-line",
			number:  PhoneNumber{CountryCode: 54, NationalNumber: 12345678901},
			numType: TypeFixedLine,
			want:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := u.IsPossibleNumberForType(tc.number, tc.numType)
			if got != tc.want {
				t.Errorf("IsPossibleNumberForType(%+v, %d) = %v, want %v",
					tc.number, tc.numType, got, tc.want)
			}
		})
	}
}


// --- GetRegionCodeForNumber ---

func TestGetRegionCodeForNumber(t *testing.T) {
	u := Instance()

	tests := []struct {
		name   string
		number PhoneNumber
		want   string
	}{
		{
			name:   "Bahamas (NANPA disambiguation)",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 2423232345},
			want:   "BS",
		},
		{
			name:   "US (NANPA)",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 6502530000},
			want:   "US",
		},
		{
			name:   "GB",
			number: PhoneNumber{CountryCode: 44, NationalNumber: 7912345678},
			want:   "GB",
		},
		{
			name:   "International toll-free (non-geo)",
			number: PhoneNumber{CountryCode: 800, NationalNumber: 12345678},
			want:   "001",
		},
		{
			name:   "Universal premium rate (non-geo)",
			number: PhoneNumber{CountryCode: 979, NationalNumber: 123456789},
			want:   "001",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := u.GetRegionCodeForNumber(tc.number)
			if got != tc.want {
				t.Errorf("GetRegionCodeForNumber(%+v) = %q, want %q",
					tc.number, got, tc.want)
			}
		})
	}
}


// --- GetCountryCodeForRegion ---

func TestGetCountryCodeForRegion(t *testing.T) {
	u := Instance()

	tests := []struct {
		region string
		want   int32
	}{
		{"US", 1},
		{"GB", 44},
		{"DE", 49},
		{"JP", 81},
		{"ZZ", 0},   // invalid
		{"", 0},     // invalid
		{"001", 0},  // non-geo is not a valid region code
	}

	for _, tc := range tests {
		t.Run(tc.region, func(t *testing.T) {
			got := u.GetCountryCodeForRegion(tc.region)
			if got != tc.want {
				t.Errorf("GetCountryCodeForRegion(%q) = %d, want %d",
					tc.region, got, tc.want)
			}
		})
	}
}

// --- Benchmarks ---

func BenchmarkIsPossibleNumber(b *testing.B) {
	u := Instance()
	number := PhoneNumber{CountryCode: 1, NationalNumber: 6502530000}
	b.ResetTimer()
	for b.Loop() {
		u.IsPossibleNumber(number)
	}
}

func BenchmarkIsValidNumber(b *testing.B) {
	u := Instance()
	number := PhoneNumber{CountryCode: 1, NationalNumber: 6502530000}
	b.ResetTimer()
	for b.Loop() {
		u.IsValidNumber(number)
	}
}
