package phonesafe

import "testing"

func FuzzFormat(f *testing.F) {
	seeds := []struct {
		cc     int32
		nn     uint64
		format int
	}{
		{1, 6502530000, 0}, // E164
		{1, 6502530000, 1}, // International
		{1, 6502530000, 2}, // National
		{1, 6502530000, 3}, // RFC3966
		{44, 2070313000, 1},
		{39, 236618300, 0},
		{81, 312345678, 2},
	}
	for _, s := range seeds {
		f.Add(s.cc, s.nn, s.format)
	}

	u := Instance()
	f.Fuzz(func(t *testing.T, cc int32, nn uint64, format int) {
		num := PhoneNumber{
			CountryCode:    cc,
			NationalNumber: nn,
		}
		fmt := PhoneNumberFormat(format % 4)
		// Format must never panic.
		u.Format(num, fmt)
	})
}

func FuzzValidate(f *testing.F) {
	seeds := []struct {
		cc int32
		nn uint64
	}{
		{1, 6502530000},
		{44, 2070313000},
		{0, 0},
		{999, 12345678901234567},
		{86, 13912345678},
	}
	for _, s := range seeds {
		f.Add(s.cc, s.nn)
	}

	u := Instance()
	f.Fuzz(func(t *testing.T, cc int32, nn uint64) {
		num := PhoneNumber{
			CountryCode:    cc,
			NationalNumber: nn,
		}
		// None of these should panic.
		u.IsValidNumber(num)
		u.IsPossibleNumber(num)
		u.IsPossibleNumberWithReason(num)
		u.GetNumberType(num)
	})
}

func FuzzFindNumbers(f *testing.F) {
	seeds := []struct {
		text   string
		region string
	}{
		{"Call +1 650 253 0000 for info.", "US"},
		{"Reach us at (650) 253-0000 or +44 20 7031 3000.", "US"},
		{"No numbers here.", "US"},
		{"+39 02 3661 8300 is the office.", "IT"},
		{"Dial 911 in emergencies.", "US"},
		{"", "US"},
		{"123", "ZZ"},
	}
	for _, s := range seeds {
		f.Add(s.text, s.region)
	}

	u := Instance()
	f.Fuzz(func(t *testing.T, text, region string) {
		// FindNumbers must never panic, regardless of input.
		m := u.FindNumbers(text, region, LeniencyPossible, 100)
		for m.HasNext() {
			_ = m.Next()
		}
	})
}

func FuzzAsYouType(f *testing.F) {
	// Seeds: representative digit sequences for various regions.
	seeds := []struct {
		digits string
		region string
	}{
		{"6502530000", "US"},
		{"16502530000", "US"},
		{"2070313000", "GB"},
		{"0312345678", "JP"},
		{"03012345678", "DE"},
		{"01155123456789", "US"}, // IDD to AR
		{"", "US"},
		{"1", "US"},
		{"12345678901234567890", "US"}, // very long
	}
	for _, s := range seeds {
		f.Add(s.digits, s.region)
	}

	u := Instance()
	f.Fuzz(func(t *testing.T, digits, region string) {
		// NewAsYouTypeFormatter + InputDigit must never panic.
		formatter := u.NewAsYouTypeFormatter(region)
		for _, ch := range digits {
			formatter.InputDigit(ch)
		}
	})
}

func FuzzParse(f *testing.F) {
	// Seed corpus with representative inputs.
	seeds := []struct {
		number string
		region string
	}{
		{"+1 650 253 0000", "US"},
		{"(650) 253-0000", "US"},
		{"+44 20 7031 3000", "GB"},
		{"044 668 1800", "CH"},
		{"+39 02 3661 8300", "IT"},
		{"tel:+1-650-253-0000;ext=123", "ZZ"},
		{"011 44 20 7031 3000", "US"},
		{"+1 800 SIX-FLAGS", "US"},
		{"", "US"},
		{"hello", "US"},
		{"+999 123 456", "ZZ"},
		{"1234567890123456789", "US"},
		{"+1 650 253 0000 ext. 123", "ZZ"},
		{"０３-１２３４-５６７８", "JP"},
	}
	for _, s := range seeds {
		f.Add(s.number, s.region)
	}

	u := Instance()
	f.Fuzz(func(t *testing.T, number, region string) {
		// Parse must never panic, regardless of input.
		u.Parse(number, region)
	})
}
