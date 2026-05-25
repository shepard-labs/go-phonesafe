package phonesafe

import "testing"

func TestNormalizeDigitsOnly(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"ascii digits", "123-456-7890", "1234567890"},
		{"with letters", "abc123def", "123"},
		{"empty", "", ""},
		{"only non-digits", "abc", ""},
		{"fullwidth digits", "０１２３", "0123"},
		{"arabic-indic", "٠١٢٣٤٥", "012345"},
		{"eastern arabic-indic", "۰۱۲۳۴", "01234"},
		{"mixed unicode", "٠1２۳4", "01234"},
		{"with punctuation", "+1 (650) 253-0000", "16502530000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeDigitsOnly(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeDigitsOnly(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeDiallableCharsOnly(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"digits and plus", "+1-650-253-0000", "+16502530000"},
		{"with star and hash", "123*456#789", "123*456#789"},
		{"strip alpha", "1800SIXFLAGS", "1800"},
		{"empty", "", ""},
		{"fullwidth digits", "０１２＋", "012+"},
		{"only symbols", "(---)", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeDiallableCharsOnly(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeDiallableCharsOnly(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConvertAlphaCharactersInNumber(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"vanity", "1-800-FLOWERS", "1-800-3569377"},
		{"lowercase", "1-800-flowers", "1-800-3569377"},
		{"mixed case", "1-800-FlOwEr", "1-800-356937"},
		{"no alpha", "1-800-123-4567", "1-800-123-4567"},
		{"all alpha", "ABCDEFGHIJKLMNOPQRSTUVWXYZ", "22233344455566677778889999"},
		{"with punctuation preserved", "(800) GOT-JUNK", "(800) 468-5865"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertAlphaCharactersInNumber(tt.input)
			if got != tt.want {
				t.Errorf("ConvertAlphaCharactersInNumber(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"vanity number", "1-800-SIX-FLAGS", "180074935247"},
		{"no alpha (under 3)", "1-800-ab-4567", "18004567"},
		{"digits only", "+1 650 253 0000", "16502530000"},
		{"three ascii alpha", "1-800-TEST", "18008378"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalize(tt.input)
			if got != tt.want {
				t.Errorf("normalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
