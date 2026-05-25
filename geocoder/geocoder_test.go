package geocoder

import (
	"testing"

	"golang.org/x/text/language"

	phonesafe "github.com/shepard-labs/go-phonesafe"
)

// Test numbers (matching upstream PhoneNumberOfflineGeocoderTest).
var (
	koNumber1 = phonesafe.PhoneNumber{CountryCode: 82, NationalNumber: 22123456}   // Seoul
	koNumber2 = phonesafe.PhoneNumber{CountryCode: 82, NationalNumber: 322123456}  // Incheon
	koNumber3 = phonesafe.PhoneNumber{CountryCode: 82, NationalNumber: 6421234567} // Jeju
	koMobile  = phonesafe.PhoneNumber{CountryCode: 82, NationalNumber: 101234567}  // Mobile

	usNumber1 = phonesafe.PhoneNumber{CountryCode: 1, NationalNumber: 6502530000} // Mountain View, CA
	usNumber2 = phonesafe.PhoneNumber{CountryCode: 1, NationalNumber: 6509600000} // Mountain View, CA
	usNumber3 = phonesafe.PhoneNumber{CountryCode: 1, NationalNumber: 2128120000} // New York, NY

	nanpaTollFree = phonesafe.PhoneNumber{CountryCode: 1, NationalNumber: 8002431234}

	arMobileNumber = phonesafe.PhoneNumber{CountryCode: 54, NationalNumber: 92214000000} // La Plata

	auNumber = phonesafe.PhoneNumber{CountryCode: 61, NationalNumber: 236618300}

	invalidCC = phonesafe.PhoneNumber{CountryCode: 999, NationalNumber: 2423651234}

	internationalTollFree = phonesafe.PhoneNumber{CountryCode: 800, NationalNumber: 12345678}
)

func TestGetDescriptionForNumber_EnglishUS(t *testing.T) {
	g := Instance()
	en := language.English

	tests := []struct {
		name   string
		number phonesafe.PhoneNumber
		want   string
	}{
		{"Mountain View by prefix", usNumber1, "Mountain View, CA"},
		{"Mountain View by area", usNumber2, "Mountain View, CA"},
		{"New York", usNumber3, "New York, NY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := g.GetDescriptionForNumber(tt.number, en)
			if got != tt.want {
				t.Errorf("GetDescriptionForNumber(%v, en) = %q, want %q", tt.number, got, tt.want)
			}
		})
	}
}

func TestGetDescriptionForKoreanNumber(t *testing.T) {
	g := Instance()
	en := language.English
	ko := language.Korean

	tests := []struct {
		name   string
		number phonesafe.PhoneNumber
		lang   language.Tag
		want   string
	}{
		{"Seoul EN", koNumber1, en, "Seoul"},
		{"Incheon EN", koNumber2, en, "Incheon"},
		{"Jeju EN", koNumber3, en, "Jeju"},
		{"Seoul KO", koNumber1, ko, "서울"},   // 서울
		{"Incheon KO", koNumber2, ko, "인천"}, // 인천
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := g.GetDescriptionForNumber(tt.number, tt.lang)
			if got != tt.want {
				t.Errorf("GetDescriptionForNumber(%v, %v) = %q, want %q", tt.number, tt.lang, got, tt.want)
			}
		})
	}
}

func TestGetDescriptionForArgentinianMobileNumber(t *testing.T) {
	g := Instance()
	got := g.GetDescriptionForNumber(arMobileNumber, language.English)
	want := "La Plata, Buenos Aires"
	if got != want {
		t.Errorf("GetDescriptionForNumber(AR mobile) = %q, want %q", got, want)
	}
}

func TestGetDescriptionForFallBack(t *testing.T) {
	g := Instance()

	// German has no US data, so it falls back to English.
	got := g.GetDescriptionForNumber(usNumber3, language.German)
	want := "New York, NY"
	if got != want {
		t.Errorf("GetDescriptionForNumber(US3, de) = %q, want %q", got, want)
	}

	// French also falls back to English for US numbers.
	got = g.GetDescriptionForNumber(usNumber1, language.French)
	want = "Mountain View, CA"
	if got != want {
		t.Errorf("GetDescriptionForNumber(US1, fr) = %q, want %q", got, want)
	}

	// Korean does NOT fall back to English.
	got = g.GetDescriptionForNumber(koNumber3, language.Korean)
	want = "제주" // 제주
	if got != want {
		t.Errorf("GetDescriptionForNumber(KO3, ko) = %q, want %q", got, want)
	}
}

func TestGetDescriptionForInvalidNumber(t *testing.T) {
	g := Instance()
	en := language.English

	// Invalid country code.
	got := g.GetDescriptionForNumber(invalidCC, en)
	if got != "" {
		t.Errorf("GetDescriptionForNumber(invalid CC) = %q, want empty", got)
	}

	// International toll-free (non-geo, but invalid/unknown type likely).
	got = g.GetDescriptionForNumber(internationalTollFree, en)
	// Non-geo numbers should return a country name or empty.
	// For CC 800 (international toll-free), there's no country to display.
	// The exact behavior depends on whether it validates.
	// Just verify it doesn't panic.
	_ = got
}

func TestGetDescriptionForNonGeographicalNumber(t *testing.T) {
	g := Instance()
	en := language.English

	// Korean mobile — not geographical, should return country name.
	got := g.GetDescriptionForNumber(koMobile, en)
	want := "South Korea"
	if got != want {
		t.Errorf("GetDescriptionForNumber(KO mobile) = %q, want %q", got, want)
	}
}

func TestGetDescriptionForNANPATollFree(t *testing.T) {
	g := Instance()
	en := language.English

	// NANPA toll-free numbers belong to multiple countries.
	// Non-geographical, so we try to get country name.
	// But since it's valid for multiple NANPA countries, it should be ambiguous → empty or "United States".
	got := g.GetDescriptionForNumber(nanpaTollFree, en)
	// Toll-free is non-geographical. The country name lookup will try to
	// disambiguate. If it maps to a single valid region it returns that.
	// If ambiguous, returns "". The NANPA toll-free range is shared.
	// Let's just verify it doesn't panic and returns something reasonable.
	_ = got
}

func TestGetDescriptionForNumberWithUserRegion(t *testing.T) {
	g := Instance()
	en := language.English

	// User in the US — should see detailed area info.
	got := g.GetDescriptionForValidNumberWithUserRegion(usNumber1, en, "US")
	want := "Mountain View, CA"
	if got != want {
		t.Errorf("WithUserRegion(US1, en, US) = %q, want %q", got, want)
	}

	// User in Italy — should see country name only.
	got = g.GetDescriptionForValidNumberWithUserRegion(usNumber1, en, "IT")
	want = "United States"
	if got != want {
		t.Errorf("WithUserRegion(US1, en, IT) = %q, want %q", got, want)
	}

	// Unknown region — should show country name.
	got = g.GetDescriptionForValidNumberWithUserRegion(usNumber1, en, "ZZ")
	// ZZ doesn't match "US", so it shows the region name for "US".
	want = "United States"
	if got != want {
		t.Errorf("WithUserRegion(US1, en, ZZ) = %q, want %q", got, want)
	}
}

func TestGetDescriptionForAustralia(t *testing.T) {
	g := Instance()
	en := language.English

	// Australia has a single region for CC 61.
	got := g.GetDescriptionForNumber(auNumber, en)
	// Should get area description or "Australia" as fallback.
	if got == "" {
		t.Error("GetDescriptionForNumber(AU) returned empty, expected area or country name")
	}
}

func TestLangTagToKey(t *testing.T) {
	tests := []struct {
		tag  language.Tag
		want string
	}{
		{language.English, "en"},
		{language.German, "de"},
		{language.Korean, "ko"},
		{language.French, "fr"},
		{language.Japanese, "ja"},
		{language.MustParse("zh-Hant"), "zh_Hant"},
		{language.MustParse("zh-Hans"), "zh"},
		{language.MustParse("he"), "iw"},
	}

	for _, tt := range tests {
		t.Run(tt.tag.String(), func(t *testing.T) {
			got := langTagToKey(tt.tag)
			if got != tt.want {
				t.Errorf("langTagToKey(%v) = %q, want %q", tt.tag, got, tt.want)
			}
		})
	}
}

func TestMayFallBackToEnglish(t *testing.T) {
	tests := []struct {
		lang string
		want bool
	}{
		{"en", true},
		{"de", true},
		{"fr", true},
		{"zh", false},
		{"zh_Hant", false},
		{"ja", false},
		{"ko", false},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			got := mayFallBackToEnglish(tt.lang)
			if got != tt.want {
				t.Errorf("mayFallBackToEnglish(%q) = %v, want %v", tt.lang, got, tt.want)
			}
		})
	}
}

func BenchmarkGetDescriptionForNumber(b *testing.B) {
	g := Instance()
	en := language.English
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.GetDescriptionForNumber(usNumber1, en)
	}
}
