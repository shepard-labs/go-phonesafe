package phonesafe

import "testing"

// --- GetNumberType ---

func TestGetNumberType_PremiumRate(t *testing.T) {
	u := Instance()

	tests := []struct {
		name   string
		number PhoneNumber
	}{
		{"US premium", PhoneNumber{CountryCode: 1, NationalNumber: 9002530000}},
		{"IT premium", PhoneNumber{CountryCode: 39, NationalNumber: 892123}},
		{"GB premium", PhoneNumber{CountryCode: 44, NationalNumber: 9187654321}},
		{"DE premium 10-digit", PhoneNumber{CountryCode: 49, NationalNumber: 9001654321}},
		{"DE premium 11-digit", PhoneNumber{CountryCode: 49, NationalNumber: 90091234567}},
		{"Universal premium rate", PhoneNumber{CountryCode: 979, NationalNumber: 123456789}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := u.GetNumberType(tc.number)
			if got != TypePremiumRate {
				t.Errorf("GetNumberType(%+v) = %v, want TypePremiumRate", tc.number, got)
			}
		})
	}
}

func TestGetNumberType_TollFree(t *testing.T) {
	u := Instance()

	tests := []struct {
		name   string
		number PhoneNumber
	}{
		{"US toll-free 800", PhoneNumber{CountryCode: 1, NationalNumber: 8002530000}},
		{"US toll-free 888", PhoneNumber{CountryCode: 1, NationalNumber: 8882345678}},
		{"IT toll-free", PhoneNumber{CountryCode: 39, NationalNumber: 803123}},
		{"GB toll-free 800", PhoneNumber{CountryCode: 44, NationalNumber: 8001234567}},
		{"DE toll-free", PhoneNumber{CountryCode: 49, NationalNumber: 8001234567}},
		{"International toll-free", PhoneNumber{CountryCode: 800, NationalNumber: 12345678}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := u.GetNumberType(tc.number)
			if got != TypeTollFree {
				t.Errorf("GetNumberType(%+v) = %v, want TypeTollFree", tc.number, got)
			}
		})
	}
}

func TestGetNumberType_Mobile(t *testing.T) {
	u := Instance()

	tests := []struct {
		name   string
		number PhoneNumber
	}{
		{"BS mobile", PhoneNumber{CountryCode: 1, NationalNumber: 2423570000}},
		{"GB mobile", PhoneNumber{CountryCode: 44, NationalNumber: 7912345678}},
		{"IT mobile", PhoneNumber{CountryCode: 39, NationalNumber: 345678901}},
		{"AR mobile", PhoneNumber{CountryCode: 54, NationalNumber: 91187654321}},
		{"DE mobile", PhoneNumber{CountryCode: 49, NationalNumber: 15123456789}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := u.GetNumberType(tc.number)
			if got != TypeMobile {
				t.Errorf("GetNumberType(%+v) = %v, want TypeMobile", tc.number, got)
			}
		})
	}
}

func TestGetNumberType_FixedLine(t *testing.T) {
	u := Instance()

	tests := []struct {
		name   string
		number PhoneNumber
	}{
		{"BS fixed-line", PhoneNumber{CountryCode: 1, NationalNumber: 2423651234}},
		{"IT fixed-line", PhoneNumber{CountryCode: 39, NationalNumber: 236618300, ItalianLeadingZero: true}},
		{"GB fixed-line", PhoneNumber{CountryCode: 44, NationalNumber: 2070313000}},
		{"DE fixed-line", PhoneNumber{CountryCode: 49, NationalNumber: 30123456}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := u.GetNumberType(tc.number)
			if got != TypeFixedLine {
				t.Errorf("GetNumberType(%+v) = %v, want TypeFixedLine", tc.number, got)
			}
		})
	}
}

func TestGetNumberType_FixedLineOrMobile(t *testing.T) {
	u := Instance()

	tests := []struct {
		name   string
		number PhoneNumber
	}{
		// US has identical FixedLine and Mobile patterns → always FixedLineOrMobile
		{"US number", PhoneNumber{CountryCode: 1, NationalNumber: 6502530000}},
		{"US number 2", PhoneNumber{CountryCode: 1, NationalNumber: 2123456789}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := u.GetNumberType(tc.number)
			if got != TypeFixedLineOrMobile {
				t.Errorf("GetNumberType(%+v) = %v, want TypeFixedLineOrMobile", tc.number, got)
			}
		})
	}
}

func TestGetNumberType_SharedCost(t *testing.T) {
	u := Instance()

	// Austria (CC=43) SharedCost pattern: 8(?:10|2[018])\d{6,10}|828\d{5}
	number := PhoneNumber{CountryCode: 43, NationalNumber: 810123456}
	got := u.GetNumberType(number)
	if got != TypeSharedCost {
		t.Errorf("GetNumberType(%+v) = %v, want TypeSharedCost", number, got)
	}
}

func TestGetNumberType_VOIP(t *testing.T) {
	u := Instance()

	// GB VOIP pattern: 56\d{8}
	number := PhoneNumber{CountryCode: 44, NationalNumber: 5612345678}
	got := u.GetNumberType(number)
	if got != TypeVOIP {
		t.Errorf("GetNumberType(%+v) = %v, want TypeVOIP", number, got)
	}
}

func TestGetNumberType_PersonalNumber(t *testing.T) {
	u := Instance()

	// GB PersonalNumber pattern: 70\d{8}
	number := PhoneNumber{CountryCode: 44, NationalNumber: 7012345678}
	got := u.GetNumberType(number)
	if got != TypePersonalNumber {
		t.Errorf("GetNumberType(%+v) = %v, want TypePersonalNumber", number, got)
	}
}

func TestGetNumberType_Unknown(t *testing.T) {
	u := Instance()

	tests := []struct {
		name   string
		number PhoneNumber
	}{
		{"US local number (too short for valid)", PhoneNumber{CountryCode: 1, NationalNumber: 2530000}},
		{"Invalid country code", PhoneNumber{CountryCode: 3923, NationalNumber: 2366}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := u.GetNumberType(tc.number)
			if got != TypeUnknown {
				t.Errorf("GetNumberType(%+v) = %v, want TypeUnknown", tc.number, got)
			}
		})
	}
}

// --- IsNumberGeographical ---

func TestIsNumberGeographical(t *testing.T) {
	u := Instance()

	tests := []struct {
		name   string
		number PhoneNumber
		want   bool
	}{
		// Geographical: fixed-line or fixed-line-or-mobile
		{
			name:   "US number (fixed-line-or-mobile) is geographical",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 6502530000},
			want:   true,
		},
		{
			name:   "AU fixed-line is geographical",
			number: PhoneNumber{CountryCode: 61, NationalNumber: 236618300},
			want:   true,
		},
		{
			name:   "DE fixed-line is geographical",
			number: PhoneNumber{CountryCode: 49, NationalNumber: 30123456},
			want:   true,
		},
		// Geo-mobile countries: mobile IS geographical
		{
			name:   "AR mobile is geographical (geo-mobile country)",
			number: PhoneNumber{CountryCode: 54, NationalNumber: 91187654321},
			want:   true,
		},
		{
			name:   "MX mobile is geographical (geo-mobile country)",
			number: PhoneNumber{CountryCode: 52, NationalNumber: 2221234567},
			want:   true,
		},
		{
			name:   "MX mobile 2 is geographical (geo-mobile country)",
			number: PhoneNumber{CountryCode: 52, NationalNumber: 5512345678},
			want:   true,
		},
		// Not geographical
		{
			name:   "BS mobile is not geographical",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 2423570000},
			want:   false,
		},
		{
			name:   "GB mobile is not geographical",
			number: PhoneNumber{CountryCode: 44, NationalNumber: 7912345678},
			want:   false,
		},
		{
			name:   "International toll-free is not geographical",
			number: PhoneNumber{CountryCode: 800, NationalNumber: 12345678},
			want:   false,
		},
		{
			name:   "US toll-free is not geographical",
			number: PhoneNumber{CountryCode: 1, NationalNumber: 8002530000},
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := u.IsNumberGeographical(tc.number)
			if got != tc.want {
				t.Errorf("IsNumberGeographical(%+v) = %v, want %v",
					tc.number, got, tc.want)
			}
		})
	}
}

// --- Benchmarks ---

func BenchmarkGetNumberType(b *testing.B) {
	u := Instance()
	number := PhoneNumber{CountryCode: 1, NationalNumber: 6502530000}
	b.ResetTimer()
	for b.Loop() {
		u.GetNumberType(number)
	}
}
