package shortnumber

import (
	"testing"

	phonesafe "github.com/shepard-labs/go-phonesafe"
)

// Helper to create a PhoneNumber with country code and national number.
func pn(cc int32, nn uint64) phonesafe.PhoneNumber {
	return phonesafe.PhoneNumber{CountryCode: cc, NationalNumber: nn}
}

func TestIsPossibleShortNumber(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number phonesafe.PhoneNumber
		want   bool
	}{
		{"FR possible", pn(33, 123456), true},
		{"FR impossible (too short)", pn(33, 9), false},
		// GB and GG share CC 44; 11001 is possible but not valid.
		{"GB/GG possible (shared CC)", pn(44, 11001), true},
		{"US 911 possible", pn(1, 911), true},
		{"unknown CC", pn(123, 911), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.IsPossibleShortNumber(tt.number)
			if got != tt.want {
				t.Errorf("IsPossibleShortNumber(%+v) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

func TestIsPossibleShortNumberForRegion(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number phonesafe.PhoneNumber
		region string
		want   bool
	}{
		{"FR possible", pn(33, 123456), "FR", true},
		{"FR impossible", pn(33, 9), "FR", false},
		{"wrong region", pn(1, 911), "GB", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.IsPossibleShortNumberForRegion(tt.number, tt.region)
			if got != tt.want {
				t.Errorf("IsPossibleShortNumberForRegion(%+v, %q) = %v, want %v",
					tt.number, tt.region, got, tt.want)
			}
		})
	}
}

func TestIsValidShortNumber(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number phonesafe.PhoneNumber
		want   bool
	}{
		{"FR 1010 valid", pn(33, 1010), true},
		{"FR 123456 invalid", pn(33, 123456), false},
		// GB and GG share CC 44; 18001 is valid in GB.
		{"GB 18001 valid (shared CC)", pn(44, 18001), true},
		{"US 911 valid", pn(1, 911), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.IsValidShortNumber(tt.number)
			if got != tt.want {
				t.Errorf("IsValidShortNumber(%+v) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

func TestIsValidShortNumberForRegion(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number phonesafe.PhoneNumber
		region string
		want   bool
	}{
		{"FR 1010 valid", pn(33, 1010), "FR", true},
		{"FR 123456 invalid", pn(33, 123456), "FR", false},
		{"US 911 valid", pn(1, 911), "US", true},
		{"wrong region", pn(1, 911), "GB", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.IsValidShortNumberForRegion(tt.number, tt.region)
			if got != tt.want {
				t.Errorf("IsValidShortNumberForRegion(%+v, %q) = %v, want %v",
					tt.number, tt.region, got, tt.want)
			}
		})
	}
}

func TestGetExpectedCostForRegion(t *testing.T) {
	info := Instance()

	// Use example numbers from metadata for FR.
	premiumExample := info.GetExampleShortNumberForCost("FR", phonesafe.CostPremiumRate)
	standardExample := info.GetExampleShortNumberForCost("FR", phonesafe.CostStandardRate)
	tollFreeExample := info.GetExampleShortNumberForCost("FR", phonesafe.CostTollFree)

	tests := []struct {
		name   string
		number phonesafe.PhoneNumber
		region string
		want   phonesafe.ShortNumberCost
	}{
		{"US 911 toll-free (emergency)", pn(1, 911), "US", phonesafe.CostTollFree},
		{"US 112 toll-free", pn(1, 112), "US", phonesafe.CostTollFree},
		{"nonexistent region", pn(1, 911), "ZZ", phonesafe.CostUnknown},
		{"unknown CC", pn(123, 911), "US", phonesafe.CostUnknown},
	}

	// Add FR cost tests if examples exist.
	if premiumExample != "" {
		nn := parseNN(premiumExample)
		tests = append(tests, struct {
			name   string
			number phonesafe.PhoneNumber
			region string
			want   phonesafe.ShortNumberCost
		}{"FR premium", pn(33, nn), "FR", phonesafe.CostPremiumRate})
	}
	if standardExample != "" {
		nn := parseNN(standardExample)
		tests = append(tests, struct {
			name   string
			number phonesafe.PhoneNumber
			region string
			want   phonesafe.ShortNumberCost
		}{"FR standard", pn(33, nn), "FR", phonesafe.CostStandardRate})
	}
	if tollFreeExample != "" {
		nn := parseNN(tollFreeExample)
		tests = append(tests, struct {
			name   string
			number phonesafe.PhoneNumber
			region string
			want   phonesafe.ShortNumberCost
		}{"FR toll-free", pn(33, nn), "FR", phonesafe.CostTollFree})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.GetExpectedCostForRegion(tt.number, tt.region)
			if got != tt.want {
				t.Errorf("GetExpectedCostForRegion(%+v, %q) = %v, want %v",
					tt.number, tt.region, got, tt.want)
			}
		})
	}
}

func TestGetExpectedCostSharedCallingCode(t *testing.T) {
	info := Instance()

	// AU and CX share CC 61. In AU: 1234=premium, 1194=standard, 733=toll-free.
	// These are not valid in CX, so GetExpectedCost sees mixed results.
	tests := []struct {
		name   string
		number phonesafe.PhoneNumber
		want   phonesafe.ShortNumberCost
	}{
		// Premium in AU, unknown in CX → PREMIUM wins.
		{"1234 premium wins", pn(61, 1234), phonesafe.CostPremiumRate},
		// Standard in AU, unknown in CX → UNKNOWN wins.
		{"1194 unknown wins", pn(61, 1194), phonesafe.CostUnknown},
		// Toll-free in AU, unknown in CX → UNKNOWN wins.
		{"733 unknown wins", pn(61, 733), phonesafe.CostUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.GetExpectedCost(tt.number)
			if got != tt.want {
				t.Errorf("GetExpectedCost(%+v) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

func TestConnectsToEmergencyNumber_US(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"911", "911", true},
		{"112", "112", true},
		{"999 not emergency", "999", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.ConnectsToEmergencyNumber(tt.number, "US")
			if got != tt.want {
				t.Errorf("ConnectsToEmergencyNumber(%q, US) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

func TestConnectsToEmergencyNumberLongNumber_US(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"9116666666 (prefix)", "9116666666", true},
		{"1126666666 (prefix)", "1126666666", true},
		{"9996666666 not emergency", "9996666666", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.ConnectsToEmergencyNumber(tt.number, "US")
			if got != tt.want {
				t.Errorf("ConnectsToEmergencyNumber(%q, US) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

func TestConnectsToEmergencyNumberWithFormatting_US(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"9-1-1", "9-1-1", true},
		{"1-1-2", "1-1-2", true},
		{"9-9-9 not emergency", "9-9-9", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.ConnectsToEmergencyNumber(tt.number, "US")
			if got != tt.want {
				t.Errorf("ConnectsToEmergencyNumber(%q, US) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

func TestConnectsToEmergencyNumberWithPlusSign_US(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number string
	}{
		{"+911", "+911"},
		{"fullwidth +911", "＋911"},
		{" +911 (leading space then plus)", " +911"},
		{"+112", "+112"},
		{"+999", "+999"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.ConnectsToEmergencyNumber(tt.number, "US")
			if got {
				t.Errorf("ConnectsToEmergencyNumber(%q, US) = true, want false", tt.number)
			}
		})
	}
}

func TestConnectsToEmergencyNumber_BR(t *testing.T) {
	info := Instance()

	// Brazil: emergency numbers work.
	if !info.ConnectsToEmergencyNumber("190", "BR") {
		t.Error("ConnectsToEmergencyNumber(190, BR) = false, want true")
	}
	// Brazil: extra digits do NOT work.
	if info.ConnectsToEmergencyNumber("9111", "BR") {
		t.Error("ConnectsToEmergencyNumber(9111, BR) = true, want false")
	}
	if info.ConnectsToEmergencyNumber("1900", "BR") {
		t.Error("ConnectsToEmergencyNumber(1900, BR) = true, want false")
	}
}

func TestConnectsToEmergencyNumber_CL(t *testing.T) {
	info := Instance()

	// Chile: emergency numbers work.
	if !info.ConnectsToEmergencyNumber("131", "CL") {
		t.Error("ConnectsToEmergencyNumber(131, CL) = false, want true")
	}
	if !info.ConnectsToEmergencyNumber("133", "CL") {
		t.Error("ConnectsToEmergencyNumber(133, CL) = false, want true")
	}
	// Chile: extra digits do NOT work.
	if info.ConnectsToEmergencyNumber("1313", "CL") {
		t.Error("ConnectsToEmergencyNumber(1313, CL) = true, want false")
	}
}

func TestIsEmergencyNumber(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number string
		region string
		want   bool
	}{
		{"US 911", "911", "US", true},
		{"US 112", "112", "US", true},
		{"US 999 not", "999", "US", false},
		// IsEmergencyNumber requires exact match — no prefix allowed.
		{"US 9116666 not exact", "9116666", "US", false},
		{"EU 112", "112", "DE", true},
		{"GB 999", "999", "GB", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.IsEmergencyNumber(tt.number, tt.region)
			if got != tt.want {
				t.Errorf("IsEmergencyNumber(%q, %q) = %v, want %v",
					tt.number, tt.region, got, tt.want)
			}
		})
	}
}

func TestIsCarrierSpecific(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number phonesafe.PhoneNumber
		want   bool
	}{
		{"US 33669 carrier-specific", pn(1, 33669), true},
		{"US 911 not carrier-specific", pn(1, 911), false},
		{"US 211 carrier-specific", pn(1, 211), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.IsCarrierSpecific(tt.number)
			if got != tt.want {
				t.Errorf("IsCarrierSpecific(%+v) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

func TestIsCarrierSpecificForRegion(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number phonesafe.PhoneNumber
		region string
		want   bool
	}{
		{"US 33669 carrier in US", pn(1, 33669), "US", true},
		{"US 211 carrier in US", pn(1, 211), "US", true},
		{"US 211 not carrier in BB", pn(1, 211), "BB", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.IsCarrierSpecificForRegion(tt.number, tt.region)
			if got != tt.want {
				t.Errorf("IsCarrierSpecificForRegion(%+v, %q) = %v, want %v",
					tt.number, tt.region, got, tt.want)
			}
		})
	}
}

func TestIsSmsServiceForRegion(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number phonesafe.PhoneNumber
		region string
		want   bool
	}{
		{"US 21234 SMS service", pn(1, 21234), "US", true},
		{"US 21234 not SMS in BB", pn(1, 21234), "BB", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.IsSmsServiceForRegion(tt.number, tt.region)
			if got != tt.want {
				t.Errorf("IsSmsServiceForRegion(%+v, %q) = %v, want %v",
					tt.number, tt.region, got, tt.want)
			}
		})
	}
}

func TestGetExampleShortNumber(t *testing.T) {
	info := Instance()

	// AD should have an example.
	if got := info.GetExampleShortNumber("AD"); got == "" {
		t.Error("GetExampleShortNumber(AD) returned empty")
	}
	// FR should have an example.
	if got := info.GetExampleShortNumber("FR"); got == "" {
		t.Error("GetExampleShortNumber(FR) returned empty")
	}
	// Invalid region returns empty.
	if got := info.GetExampleShortNumber("ZZ"); got != "" {
		t.Errorf("GetExampleShortNumber(ZZ) = %q, want empty", got)
	}
	if got := info.GetExampleShortNumber(""); got != "" {
		t.Errorf("GetExampleShortNumber('') = %q, want empty", got)
	}
}

func TestGetExampleShortNumberForCost(t *testing.T) {
	info := Instance()

	// FR should have examples for each cost.
	if got := info.GetExampleShortNumberForCost("FR", phonesafe.CostTollFree); got == "" {
		t.Error("GetExampleShortNumberForCost(FR, TollFree) returned empty")
	}
	if got := info.GetExampleShortNumberForCost("FR", phonesafe.CostStandardRate); got == "" {
		t.Error("GetExampleShortNumberForCost(FR, StandardRate) returned empty")
	}
	if got := info.GetExampleShortNumberForCost("FR", phonesafe.CostPremiumRate); got == "" {
		t.Error("GetExampleShortNumberForCost(FR, PremiumRate) returned empty")
	}
	// Unknown cost returns empty.
	if got := info.GetExampleShortNumberForCost("FR", phonesafe.CostUnknown); got != "" {
		t.Errorf("GetExampleShortNumberForCost(FR, Unknown) = %q, want empty", got)
	}
}

// parseNN converts a short number string to uint64.
func parseNN(s string) uint64 {
	var n uint64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + uint64(c-'0')
		}
	}
	return n
}
