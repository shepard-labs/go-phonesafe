package data

import (
	"sort"
	"testing"
)

func TestGeoDataLoaded(t *testing.T) {
	if len(GeoData) == 0 {
		t.Fatal("GeoData is empty")
	}
	// English should be present with many entries.
	en, ok := GeoData["en"]
	if !ok {
		t.Fatal("GeoData missing English (en)")
	}
	if len(en) < 1000 {
		t.Errorf("expected 1000+ English geo entries, got %d", len(en))
	}
	// Should be sorted by prefix.
	if !sort.SliceIsSorted(en, func(i, j int) bool {
		return en[i].Prefix < en[j].Prefix
	}) {
		t.Error("English geo entries not sorted by prefix")
	}
}

func TestTimezoneDataLoaded(t *testing.T) {
	if len(TimezoneData) == 0 {
		t.Fatal("TimezoneData is empty")
	}
	// Should be sorted by prefix.
	if !sort.SliceIsSorted(TimezoneData, func(i, j int) bool {
		return TimezoneData[i].Prefix < TimezoneData[j].Prefix
	}) {
		t.Error("TimezoneData not sorted by prefix")
	}
	// Prefix 1201 (Jersey City, NJ) should map to America/New_York.
	for _, e := range TimezoneData {
		if e.Prefix == 1201 {
			if len(e.TimeZones) == 0 {
				t.Error("1201 has no timezones")
			}
			if e.TimeZones[0] != "America/New_York" {
				t.Errorf("1201 timezone = %q, want America/New_York", e.TimeZones[0])
			}
			return
		}
	}
	t.Error("prefix 1201 not found in TimezoneData")
}
