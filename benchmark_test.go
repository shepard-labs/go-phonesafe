package phonesafe

import "testing"

// --- Parse ---

func BenchmarkParse(b *testing.B) {
	u := Instance()
	for b.Loop() {
		u.Parse("+1 650 253 0000", "US")
	}
}

func BenchmarkParseInternational(b *testing.B) {
	u := Instance()
	for b.Loop() {
		u.Parse("+44 20 7031 3000", "ZZ")
	}
}

// --- Validation ---

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

func BenchmarkGetNumberType(b *testing.B) {
	u := Instance()
	number := PhoneNumber{CountryCode: 1, NationalNumber: 6502530000}
	b.ResetTimer()
	for b.Loop() {
		u.GetNumberType(number)
	}
}

// --- Format ---

func BenchmarkFormatE164(b *testing.B) {
	u := Instance()
	num, _ := u.Parse("+1 650 253 0000", "US")
	for b.Loop() {
		u.Format(num, FormatE164)
	}
}

func BenchmarkFormatInternational(b *testing.B) {
	u := Instance()
	num, _ := u.Parse("+1 650 253 0000", "US")
	for b.Loop() {
		u.Format(num, FormatInternational)
	}
}

func BenchmarkFormatNational(b *testing.B) {
	u := Instance()
	num, _ := u.Parse("(650) 253-0000", "US")
	for b.Loop() {
		u.Format(num, FormatNational)
	}
}

// --- NumberMatch ---

func BenchmarkIsNumberMatch(b *testing.B) {
	u := Instance()
	n1, _ := u.Parse("+1 650 253 0000", "US")
	n2, _ := u.Parse("(650) 253-0000", "US")
	for b.Loop() {
		u.IsNumberMatch(n1, n2)
	}
}

// --- FindNumbers ---

func BenchmarkFindNumbersSingle(b *testing.B) {
	u := Instance()
	text := "Call us at +1 650 253 0000 for more info."
	for b.Loop() {
		m := u.FindNumbers(text, "US", LeniencyValid, 100)
		for m.HasNext() {
			_ = m.Next()
		}
	}
}

func BenchmarkFindNumbersMultiple(b *testing.B) {
	u := Instance()
	text := "Call us at +1 650 253 0000 or +44 20 7031 3000 for more info. " +
		"Also try +39 02 3661 8300 or +81 3 1234 5678."
	for b.Loop() {
		m := u.FindNumbers(text, "US", LeniencyValid, 100)
		for m.HasNext() {
			_ = m.Next()
		}
	}
}

// --- AsYouType ---

func BenchmarkAsYouTypeFormatter(b *testing.B) {
	u := Instance()
	for b.Loop() {
		f := u.NewAsYouTypeFormatter("US")
		for _, ch := range "6502530000" {
			f.InputDigit(ch)
		}
	}
}
