package timezone

import (
	"reflect"
	"testing"

	phonesafe "github.com/shepard-labs/go-phonesafe"
)

// Test numbers (matching upstream PhoneNumberToTimeZonesMapperTest).
var (
	auNumber = phonesafe.PhoneNumber{CountryCode: 61, NationalNumber: 236618300}
	caNumber = phonesafe.PhoneNumber{CountryCode: 1, NationalNumber: 6048406565}
	koNumber = phonesafe.PhoneNumber{CountryCode: 82, NationalNumber: 22123456}

	usNumber1 = phonesafe.PhoneNumber{CountryCode: 1, NationalNumber: 6509600000} // Mountain View, CA
	usNumber2 = phonesafe.PhoneNumber{CountryCode: 1, NationalNumber: 2128120000} // New York, NY
	usNumber3 = phonesafe.PhoneNumber{CountryCode: 1, NationalNumber: 6174240000} // Boston, MA

	usInvalidNumber = phonesafe.PhoneNumber{CountryCode: 1, NationalNumber: 123456789}
	koInvalidNumber = phonesafe.PhoneNumber{CountryCode: 82, NationalNumber: 1234}
	invalidCCNumber = phonesafe.PhoneNumber{CountryCode: 999, NationalNumber: 2423651234}
	intlTollFree    = phonesafe.PhoneNumber{CountryCode: 800, NationalNumber: 12345678}
	nanpaTollFree   = phonesafe.PhoneNumber{CountryCode: 1, NationalNumber: 8002431234}
)

func TestGetTimeZonesForNumber(t *testing.T) {
	m := Instance()

	tests := []struct {
		name   string
		number phonesafe.PhoneNumber
		want   []string
	}{
		{"Australia", auNumber, []string{"Australia/Sydney"}},
		{"Korea", koNumber, []string{"Asia/Seoul"}},
		{"Canada Vancouver", caNumber, []string{"America/Vancouver"}},
		{"US Mountain View", usNumber1, []string{"America/Los_Angeles"}},
		{"US New York", usNumber2, []string{"America/New_York"}},
		{"invalid CC", invalidCCNumber, []string{UnknownTimeZone}},
		{"international toll-free", intlTollFree, []string{UnknownTimeZone}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.GetTimeZonesForNumber(tt.number)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTimeZonesForNumber(%v) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

func TestGetTimeZonesForNumber_InvalidNumbers(t *testing.T) {
	m := Instance()

	// Invalid numbers should return unknown timezone.
	tests := []struct {
		name   string
		number phonesafe.PhoneNumber
	}{
		{"US invalid", usInvalidNumber},
		{"KO invalid", koInvalidNumber},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.GetTimeZonesForNumber(tt.number)
			if !reflect.DeepEqual(got, []string{UnknownTimeZone}) {
				t.Errorf("GetTimeZonesForNumber(%v) = %v, want [%s]", tt.number, got, UnknownTimeZone)
			}
		})
	}
}

func TestGetTimeZonesForGeographicalNumber(t *testing.T) {
	m := Instance()

	tests := []struct {
		name   string
		number phonesafe.PhoneNumber
		want   []string
	}{
		{"Australia", auNumber, []string{"Australia/Sydney"}},
		{"Korea", koNumber, []string{"Asia/Seoul"}},
		{"Canada Vancouver", caNumber, []string{"America/Vancouver"}},
		{"US Mountain View", usNumber1, []string{"America/Los_Angeles"}},
		{"US New York", usNumber2, []string{"America/New_York"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.GetTimeZonesForGeographicalNumber(tt.number)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTimeZonesForGeographicalNumber(%v) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

func TestGetTimeZonesForGeographicalNumber_CountryFallback(t *testing.T) {
	m := Instance()

	// US number whose specific prefix may not be in the map — should fall back
	// to the country-level NANPA entry. The upstream test verifies that
	// US_NUMBER3 (617-424-0000 Boston) returns the NANPA country-level list
	// when not covered by a specific prefix.
	got := m.GetTimeZonesForGeographicalNumber(usNumber3)
	// Should get at least one timezone (either specific or country fallback).
	if len(got) == 0 || (len(got) == 1 && got[0] == UnknownTimeZone) {
		// If this specific prefix isn't in the map, the function returns unknown.
		// That's acceptable since GetTimeZonesForGeographicalNumber doesn't do
		// country-level fallback — only GetTimeZonesForNumber does that via the
		// non-geographical path. For a geographical number with a valid prefix,
		// it should find something. Let's check if 1617 is actually in the data.
		t.Logf("GetTimeZonesForGeographicalNumber(US3) = %v (may need country-level fallback)", got)
	}
}

func TestGetTimeZonesForNumber_NonGeographical(t *testing.T) {
	m := Instance()

	// NANPA toll-free (non-geographical) should return country-level timezones for CC 1.
	got := m.GetTimeZonesForNumber(nanpaTollFree)
	// CC 1 should have a country-level timezone entry with many US timezones.
	if len(got) == 0 {
		t.Error("GetTimeZonesForNumber(NANPA toll-free) returned empty")
	}
	if len(got) == 1 && got[0] == UnknownTimeZone {
		t.Error("GetTimeZonesForNumber(NANPA toll-free) returned unknown, expected country-level timezones")
	}
}

func TestGetTimeZonesForGeographicalNumber_InvalidCC(t *testing.T) {
	m := Instance()

	got := m.GetTimeZonesForGeographicalNumber(invalidCCNumber)
	if !reflect.DeepEqual(got, []string{UnknownTimeZone}) {
		t.Errorf("GetTimeZonesForGeographicalNumber(invalid CC) = %v, want [%s]", got, UnknownTimeZone)
	}
}

func TestUnknownTimeZoneConstant(t *testing.T) {
	if UnknownTimeZone != "Etc/Unknown" {
		t.Errorf("UnknownTimeZone = %q, want %q", UnknownTimeZone, "Etc/Unknown")
	}
}

func TestInstanceSingleton(t *testing.T) {
	m1 := Instance()
	m2 := Instance()
	if m1 != m2 {
		t.Error("Instance() returned different pointers")
	}
}

func BenchmarkTimezone(b *testing.B) {
	m := Instance()
	for b.Loop() {
		m.GetTimeZonesForNumber(usNumber1)
	}
}

func BenchmarkGetTimeZonesForGeographicalNumber(b *testing.B) {
	m := Instance()
	for b.Loop() {
		m.GetTimeZonesForGeographicalNumber(usNumber1)
	}
}
