package phonesafe

import "testing"

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
