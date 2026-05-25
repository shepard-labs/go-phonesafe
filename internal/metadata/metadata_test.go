package metadata

import "testing"

func TestRegionMetadataLoaded(t *testing.T) {
	if len(RegionMetadata) == 0 {
		t.Fatal("RegionMetadata is empty")
	}
	// Spot check a few well-known regions.
	for _, region := range []string{"US", "GB", "DE", "JP", "AU", "CH"} {
		if _, ok := RegionMetadata[region]; !ok {
			t.Errorf("RegionMetadata missing region %q", region)
		}
	}
}

func TestUSMetadata(t *testing.T) {
	us := RegionMetadata["US"]
	if us == nil {
		t.Fatal("US metadata is nil")
	}
	if us.CountryCode != 1 {
		t.Errorf("US country code = %d, want 1", us.CountryCode)
	}
	if us.InternationalPrefix == "" {
		t.Error("US international prefix is empty")
	}
	if us.GeneralDesc.NationalNumberPattern == "" {
		t.Error("US general desc pattern is empty")
	}
	if len(us.NumberFormats) == 0 {
		t.Error("US has no number formats")
	}
	if !us.MainCountryForCode {
		t.Error("US should be main country for code 1")
	}
}

func TestNonGeoMetadataLoaded(t *testing.T) {
	if len(NonGeoMetadata) == 0 {
		t.Fatal("NonGeoMetadata is empty")
	}
	// International Freephone (800) should be present.
	if _, ok := NonGeoMetadata[800]; !ok {
		t.Error("NonGeoMetadata missing country code 800 (International Freephone)")
	}
}

func TestCountryCodeToRegions(t *testing.T) {
	if len(CountryCodeToRegions) == 0 {
		t.Fatal("CountryCodeToRegions is empty")
	}
	// Country code 1 should have US first (main country).
	regions := CountryCodeToRegions[1]
	if len(regions) == 0 {
		t.Fatal("no regions for country code 1")
	}
	if regions[0] != "US" {
		t.Errorf("first region for cc 1 = %q, want US", regions[0])
	}
	// Should have many NANPA regions.
	if len(regions) < 20 {
		t.Errorf("expected 20+ NANPA regions, got %d", len(regions))
	}
}

func TestNANPARegions(t *testing.T) {
	if len(NANPARegions) == 0 {
		t.Fatal("NANPARegions is empty")
	}
	if _, ok := NANPARegions["US"]; !ok {
		t.Error("US not in NANPARegions")
	}
	if _, ok := NANPARegions["CA"]; !ok {
		t.Error("CA not in NANPARegions")
	}
}

func TestShortMetadataLoaded(t *testing.T) {
	if len(ShortMetadata) == 0 {
		t.Fatal("ShortMetadata is empty")
	}
	us := ShortMetadata["US"]
	if us == nil {
		t.Fatal("US short metadata is nil")
	}
	if us.GeneralDesc.NationalNumberPattern == "" {
		t.Error("US short general desc pattern is empty")
	}
	if us.Emergency.NationalNumberPattern == "" {
		t.Error("US short emergency pattern is empty")
	}
}

func TestAlternateFormatsLoaded(t *testing.T) {
	if len(AlternateFormats) == 0 {
		t.Fatal("AlternateFormats is empty")
	}
}
