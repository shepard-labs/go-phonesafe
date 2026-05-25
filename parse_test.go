package phonesafe

import "testing"

func TestParseInternationalFormat(t *testing.T) {
	u := Instance()
	tests := []struct {
		name         string
		input        string
		region       string
		wantCC       int32
		wantNational uint64
	}{
		{"US", "+1 650 253 0000", "ZZ", 1, 6502530000},
		{"US with region", "+1 650 253 0000", "US", 1, 6502530000},
		{"UK", "+44 20 7031 3000", "ZZ", 44, 2070313000},
		{"DE", "+49 89 1234567", "ZZ", 49, 891234567},
		{"JP", "+81 3 6384 9000", "ZZ", 81, 363849000},
		{"AU", "+61 2 1234 5678", "ZZ", 61, 212345678},
		{"BR", "+55 11 2345 6789", "ZZ", 55, 1123456789},
		{"CH", "+41 44 668 1800", "ZZ", 41, 446681800},
		{"NZ", "+64 3 331 6005", "ZZ", 64, 33316005},
		{"fullwidth plus", "＋1 650 253 0000", "ZZ", 1, 6502530000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pn, err := u.Parse(tt.input, tt.region)
			if err != nil {
				t.Fatalf("Parse(%q, %q) error: %v", tt.input, tt.region, err)
			}
			if pn.CountryCode != tt.wantCC {
				t.Errorf("CountryCode = %d, want %d", pn.CountryCode, tt.wantCC)
			}
			if pn.NationalNumber != tt.wantNational {
				t.Errorf("NationalNumber = %d, want %d", pn.NationalNumber, tt.wantNational)
			}
		})
	}
}

func TestParseNationalFormat(t *testing.T) {
	u := Instance()
	tests := []struct {
		name         string
		input        string
		region       string
		wantCC       int32
		wantNational uint64
	}{
		{"US", "(650) 253-0000", "US", 1, 6502530000},
		{"US dot format", "650.253.0000", "US", 1, 6502530000},
		{"NZ", "03 331 6005", "NZ", 64, 33316005},
		{"NZ with trunk", "033316005", "NZ", 64, 33316005},
		{"UK", "020 7031 3000", "GB", 44, 2070313000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pn, err := u.Parse(tt.input, tt.region)
			if err != nil {
				t.Fatalf("Parse(%q, %q) error: %v", tt.input, tt.region, err)
			}
			if pn.CountryCode != tt.wantCC {
				t.Errorf("CountryCode = %d, want %d", pn.CountryCode, tt.wantCC)
			}
			if pn.NationalNumber != tt.wantNational {
				t.Errorf("NationalNumber = %d, want %d", pn.NationalNumber, tt.wantNational)
			}
		})
	}
}

func TestParseWithExtension(t *testing.T) {
	u := Instance()
	tests := []struct {
		name         string
		input        string
		region       string
		wantCC       int32
		wantNational uint64
		wantExt      string
	}{
		{"ext.", "+1 650 253 0000 ext. 123", "ZZ", 1, 6502530000, "123"},
		{"extn", "+1 650 253 0000 extn 456", "ZZ", 1, 6502530000, "456"},
		{"x", "+1 650 253 0000 x789", "ZZ", 1, 6502530000, "789"},
		{"hash", "+1 650 253 0000-789#", "ZZ", 1, 6502530000, "789"},
		{"rfc3966 ext", "tel:+1-650-253-0000;ext=123", "ZZ", 1, 6502530000, "123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pn, err := u.Parse(tt.input, tt.region)
			if err != nil {
				t.Fatalf("Parse(%q, %q) error: %v", tt.input, tt.region, err)
			}
			if pn.CountryCode != tt.wantCC {
				t.Errorf("CountryCode = %d, want %d", pn.CountryCode, tt.wantCC)
			}
			if pn.NationalNumber != tt.wantNational {
				t.Errorf("NationalNumber = %d, want %d", pn.NationalNumber, tt.wantNational)
			}
			if pn.Extension != tt.wantExt {
				t.Errorf("Extension = %q, want %q", pn.Extension, tt.wantExt)
			}
		})
	}
}

func TestParseItalianLeadingZeros(t *testing.T) {
	u := Instance()
	tests := []struct {
		name            string
		input           string
		region          string
		wantCC          int32
		wantNational    uint64
		wantLeadingZero bool
		wantNumZeros    int32
	}{
		{"IT fixed", "+39 02 3661 8300", "ZZ", 39, 236618300, true, 0},
		{"IT mobile (no zero)", "+39 345 678 9012", "ZZ", 39, 3456789012, false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pn, err := u.Parse(tt.input, tt.region)
			if err != nil {
				t.Fatalf("Parse(%q, %q) error: %v", tt.input, tt.region, err)
			}
			if pn.CountryCode != tt.wantCC {
				t.Errorf("CountryCode = %d, want %d", pn.CountryCode, tt.wantCC)
			}
			if pn.NationalNumber != tt.wantNational {
				t.Errorf("NationalNumber = %d, want %d", pn.NationalNumber, tt.wantNational)
			}
			if pn.ItalianLeadingZero != tt.wantLeadingZero {
				t.Errorf("ItalianLeadingZero = %v, want %v", pn.ItalianLeadingZero, tt.wantLeadingZero)
			}
			if tt.wantNumZeros != 0 && pn.NumberOfLeadingZeros != tt.wantNumZeros {
				t.Errorf("NumberOfLeadingZeros = %d, want %d", pn.NumberOfLeadingZeros, tt.wantNumZeros)
			}
		})
	}
}

func TestParseNANPA(t *testing.T) {
	u := Instance()
	tests := []struct {
		name         string
		input        string
		region       string
		wantCC       int32
		wantNational uint64
	}{
		{"US with 1", "1 650 253 0000", "US", 1, 6502530000},
		{"US without 1", "650 253 0000", "US", 1, 6502530000},
		{"US parentheses", "(650) 253-0000", "US", 1, 6502530000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pn, err := u.Parse(tt.input, tt.region)
			if err != nil {
				t.Fatalf("Parse(%q, %q) error: %v", tt.input, tt.region, err)
			}
			if pn.CountryCode != tt.wantCC {
				t.Errorf("CountryCode = %d, want %d", pn.CountryCode, tt.wantCC)
			}
			if pn.NationalNumber != tt.wantNational {
				t.Errorf("NationalNumber = %d, want %d", pn.NationalNumber, tt.wantNational)
			}
		})
	}
}

func TestParseRFC3966(t *testing.T) {
	u := Instance()
	tests := []struct {
		name         string
		input        string
		region       string
		wantCC       int32
		wantNational uint64
		wantExt      string
	}{
		{"basic", "tel:+1-650-253-0000", "ZZ", 1, 6502530000, ""},
		{"with ext", "tel:+1-650-253-0000;ext=123", "ZZ", 1, 6502530000, "123"},
		{"with phone-context global", "tel:0650-253-0000;phone-context=+64", "ZZ", 64, 6502530000, ""},
		{"with isdn subaddress", "tel:+1-650-253-0000;isub=user@example.com", "ZZ", 1, 6502530000, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pn, err := u.Parse(tt.input, tt.region)
			if err != nil {
				t.Fatalf("Parse(%q, %q) error: %v", tt.input, tt.region, err)
			}
			if pn.CountryCode != tt.wantCC {
				t.Errorf("CountryCode = %d, want %d", pn.CountryCode, tt.wantCC)
			}
			if pn.NationalNumber != tt.wantNational {
				t.Errorf("NationalNumber = %d, want %d", pn.NationalNumber, tt.wantNational)
			}
			if pn.Extension != tt.wantExt {
				t.Errorf("Extension = %q, want %q", pn.Extension, tt.wantExt)
			}
		})
	}
}

func TestParseWithIDD(t *testing.T) {
	u := Instance()
	tests := []struct {
		name         string
		input        string
		region       string
		wantCC       int32
		wantNational uint64
	}{
		{"US IDD to UK", "011 44 20 7031 3000", "US", 44, 2070313000},
		{"UK IDD to US", "00 1 650 253 0000", "GB", 1, 6502530000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pn, err := u.Parse(tt.input, tt.region)
			if err != nil {
				t.Fatalf("Parse(%q, %q) error: %v", tt.input, tt.region, err)
			}
			if pn.CountryCode != tt.wantCC {
				t.Errorf("CountryCode = %d, want %d", pn.CountryCode, tt.wantCC)
			}
			if pn.NationalNumber != tt.wantNational {
				t.Errorf("NationalNumber = %d, want %d", pn.NationalNumber, tt.wantNational)
			}
		})
	}
}

func TestParseAlphaNumbers(t *testing.T) {
	u := Instance()
	tests := []struct {
		name         string
		input        string
		region       string
		wantCC       int32
		wantNational uint64
	}{
		{"vanity US", "+1 800 SIX-FLAGS", "US", 1, 80074935247},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pn, err := u.Parse(tt.input, tt.region)
			if err != nil {
				t.Fatalf("Parse(%q, %q) error: %v", tt.input, tt.region, err)
			}
			if pn.CountryCode != tt.wantCC {
				t.Errorf("CountryCode = %d, want %d", pn.CountryCode, tt.wantCC)
			}
			if pn.NationalNumber != tt.wantNational {
				t.Errorf("NationalNumber = %d, want %d", pn.NationalNumber, tt.wantNational)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	u := Instance()
	tests := []struct {
		name     string
		input    string
		region   string
		wantCode ErrorCode
	}{
		{"not a number - empty", "", "US", ErrNotANumber},
		{"not a number - letters only", "hello", "US", ErrNotANumber},
		{"not a number - symbols", "+++", "US", ErrNotANumber},
		{"invalid country code", "+999 123 456", "ZZ", ErrInvalidCountryCode},
		{"too short after IDD", "011 1", "US", ErrTooShortAfterIDD},
		{"too short NSN", "+44 2", "ZZ", ErrTooShortNSN},
		{"too long NSN", "+1 1234567890123456789", "ZZ", ErrTooLongNSN},
		{"invalid region no plus", "650 253 0000", "ZZ", ErrInvalidCountryCode},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := u.Parse(tt.input, tt.region)
			if err == nil {
				t.Fatalf("Parse(%q, %q) expected error, got nil", tt.input, tt.region)
			}
			pe, ok := err.(*ParseError)
			if !ok {
				t.Fatalf("Parse(%q, %q) error is not *ParseError: %T", tt.input, tt.region, err)
			}
			if pe.Code != tt.wantCode {
				t.Errorf("error code = %d, want %d (error: %v)", pe.Code, tt.wantCode, pe)
			}
		})
	}
}

func TestParseAndKeepRawInput(t *testing.T) {
	u := Instance()
	pn, err := u.ParseAndKeepRawInput("+1 650 253 0000", "US")
	if err != nil {
		t.Fatalf("ParseAndKeepRawInput error: %v", err)
	}
	if pn.RawInput != "+1 650 253 0000" {
		t.Errorf("RawInput = %q, want %q", pn.RawInput, "+1 650 253 0000")
	}
	if pn.CountryCodeSource != CountryCodeFromNumberWithPlus {
		t.Errorf("CountryCodeSource = %d, want %d", pn.CountryCodeSource, CountryCodeFromNumberWithPlus)
	}
}

func TestGetNationalSignificantNumber(t *testing.T) {
	tests := []struct {
		name string
		pn   PhoneNumber
		want string
	}{
		{
			"US",
			PhoneNumber{CountryCode: 1, NationalNumber: 6502530000},
			"6502530000",
		},
		{
			"IT with leading zero",
			PhoneNumber{CountryCode: 39, NationalNumber: 236618300, ItalianLeadingZero: true},
			"0236618300",
		},
		{
			"multiple leading zeros",
			PhoneNumber{CountryCode: 39, NationalNumber: 123, ItalianLeadingZero: true, NumberOfLeadingZeros: 3},
			"000123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetNationalSignificantNumber(tt.pn)
			if got != tt.want {
				t.Errorf("GetNationalSignificantNumber = %q, want %q", got, tt.want)
			}
		})
	}
}
