package shortnumber

import (
	"testing"

	phonesafe "github.com/shepard-labs/go-phonesafe"
)

// Helper to create a PhoneNumber with country code and national number.
func pn(cc int32, nn uint64) phonesafe.PhoneNumber {
	return phonesafe.PhoneNumber{CountryCode: cc, NationalNumber: nn}
}

// Upstream: ShortNumberInfoTest.testIsPossibleShortNumber
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

// Upstream: ShortNumberInfoTest.testIsPossibleShortNumber
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

// Upstream: ShortNumberInfoTest.testIsValidShortNumber
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

// Upstream: ShortNumberInfoTest.testIsValidShortNumber
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

// Upstream: ShortNumberInfoTest.testGetExpectedCost
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

// Upstream: ShortNumberInfoTest.testGetExpectedCostForSharedCountryCallingCode
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

// Upstream: ShortNumberInfoTest.testConnectsToEmergencyNumber_US
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

// Upstream: ShortNumberInfoTest.testConnectsToEmergencyNumberLongNumber_US
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

// Upstream: ShortNumberInfoTest.testConnectsToEmergencyNumberWithFormatting_US
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

// Upstream: ShortNumberInfoTest.testConnectsToEmergencyNumberWithPlusSign_US
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

// Upstream: ShortNumberInfoTest.testConnectsToEmergencyNumber_BR
func TestConnectsToEmergencyNumber_BR(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"911 emergency", "911", true},
		{"190 emergency", "190", true},
		{"999 not emergency", "999", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.ConnectsToEmergencyNumber(tt.number, "BR")
			if got != tt.want {
				t.Errorf("ConnectsToEmergencyNumber(%q, BR) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

// Upstream: ShortNumberInfoTest.testConnectsToEmergencyNumberLongNumber_BR
func TestConnectsToEmergencyNumberLongNumber_BR(t *testing.T) {
	info := Instance()

	// Brazilian emergency numbers don't work when additional digits are appended.
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"9111 not emergency", "9111", false},
		{"1900 not emergency", "1900", false},
		{"9996 not emergency", "9996", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.ConnectsToEmergencyNumber(tt.number, "BR")
			if got != tt.want {
				t.Errorf("ConnectsToEmergencyNumber(%q, BR) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

// Upstream: ShortNumberInfoTest.testConnectsToEmergencyNumber_CL
func TestConnectsToEmergencyNumber_CL(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"131 emergency", "131", true},
		{"133 emergency", "133", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.ConnectsToEmergencyNumber(tt.number, "CL")
			if got != tt.want {
				t.Errorf("ConnectsToEmergencyNumber(%q, CL) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

// Upstream: ShortNumberInfoTest.testConnectsToEmergencyNumberLongNumber_CL
func TestConnectsToEmergencyNumberLongNumber_CL(t *testing.T) {
	info := Instance()

	// Chilean emergency numbers don't work when additional digits are appended.
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"1313 not emergency", "1313", false},
		{"1330 not emergency", "1330", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.ConnectsToEmergencyNumber(tt.number, "CL")
			if got != tt.want {
				t.Errorf("ConnectsToEmergencyNumber(%q, CL) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

// Upstream: ShortNumberInfoTest.testConnectsToEmergencyNumber_AO
func TestConnectsToEmergencyNumber_AO(t *testing.T) {
	info := Instance()

	// Angola doesn't have any metadata for emergency numbers in the test metadata.
	tests := []struct {
		name   string
		number string
	}{
		{"911", "911"},
		{"222123456", "222123456"},
		{"923123456", "923123456"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if info.ConnectsToEmergencyNumber(tt.number, "AO") {
				t.Errorf("ConnectsToEmergencyNumber(%q, AO) = true, want false", tt.number)
			}
		})
	}
}

// Upstream: ShortNumberInfoTest.testConnectsToEmergencyNumber_ZW
func TestConnectsToEmergencyNumber_ZW(t *testing.T) {
	info := Instance()

	// Zimbabwe doesn't have any metadata in the test metadata.
	tests := []struct {
		name   string
		number string
	}{
		{"911", "911"},
		{"01312345", "01312345"},
		{"0711234567", "0711234567"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if info.ConnectsToEmergencyNumber(tt.number, "ZW") {
				t.Errorf("ConnectsToEmergencyNumber(%q, ZW) = true, want false", tt.number)
			}
		})
	}
}

// Upstream: ShortNumberInfoTest.testIsEmergencyNumber_US
func TestIsEmergencyNumber_US(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"911", "911", true},
		{"112", "112", true},
		{"999 not", "999", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.IsEmergencyNumber(tt.number, "US")
			if got != tt.want {
				t.Errorf("IsEmergencyNumber(%q, US) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

// Upstream: ShortNumberInfoTest.testIsEmergencyNumberLongNumber_US
func TestIsEmergencyNumberLongNumber_US(t *testing.T) {
	info := Instance()

	// IsEmergencyNumber requires exact match — no prefix allowed.
	tests := []struct {
		name   string
		number string
	}{
		{"9116666666", "9116666666"},
		{"1126666666", "1126666666"},
		{"9996666666", "9996666666"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if info.IsEmergencyNumber(tt.number, "US") {
				t.Errorf("IsEmergencyNumber(%q, US) = true, want false", tt.number)
			}
		})
	}
}

// Upstream: ShortNumberInfoTest.testIsEmergencyNumberWithFormatting_US
func TestIsEmergencyNumberWithFormatting_US(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"9-1-1", "9-1-1", true},
		{"*911", "*911", true},
		{"1-1-2", "1-1-2", true},
		{"*112", "*112", true},
		{"9-9-9 not", "9-9-9", false},
		{"*999 not", "*999", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.IsEmergencyNumber(tt.number, "US")
			if got != tt.want {
				t.Errorf("IsEmergencyNumber(%q, US) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

// Upstream: ShortNumberInfoTest.testIsEmergencyNumberWithPlusSign_US
func TestIsEmergencyNumberWithPlusSign_US(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number string
	}{
		{"+911", "+911"},
		{"fullwidth +911", "＋911"},
		{" +911", " +911"},
		{"+112", "+112"},
		{"+999", "+999"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if info.IsEmergencyNumber(tt.number, "US") {
				t.Errorf("IsEmergencyNumber(%q, US) = true, want false", tt.number)
			}
		})
	}
}

// Upstream: ShortNumberInfoTest.testIsEmergencyNumber_BR
func TestIsEmergencyNumber_BR(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"911", "911", true},
		{"190", "190", true},
		{"999 not", "999", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info.IsEmergencyNumber(tt.number, "BR")
			if got != tt.want {
				t.Errorf("IsEmergencyNumber(%q, BR) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

// Upstream: ShortNumberInfoTest.testIsEmergencyNumberLongNumber_BR
func TestIsEmergencyNumberLongNumber_BR(t *testing.T) {
	info := Instance()

	tests := []struct {
		name   string
		number string
	}{
		{"9111", "9111"},
		{"1900", "1900"},
		{"9996", "9996"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if info.IsEmergencyNumber(tt.number, "BR") {
				t.Errorf("IsEmergencyNumber(%q, BR) = true, want false", tt.number)
			}
		})
	}
}

// Upstream: ShortNumberInfoTest.testIsEmergencyNumber_AO
func TestIsEmergencyNumber_AO(t *testing.T) {
	info := Instance()

	// Angola doesn't have any metadata for emergency numbers in the test metadata.
	tests := []struct {
		name   string
		number string
	}{
		{"911", "911"},
		{"222123456", "222123456"},
		{"923123456", "923123456"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if info.IsEmergencyNumber(tt.number, "AO") {
				t.Errorf("IsEmergencyNumber(%q, AO) = true, want false", tt.number)
			}
		})
	}
}

// Upstream: ShortNumberInfoTest.testIsEmergencyNumber_ZW
func TestIsEmergencyNumber_ZW(t *testing.T) {
	info := Instance()

	// Zimbabwe doesn't have any metadata in the test metadata.
	tests := []struct {
		name   string
		number string
	}{
		{"911", "911"},
		{"01312345", "01312345"},
		{"0711234567", "0711234567"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if info.IsEmergencyNumber(tt.number, "ZW") {
				t.Errorf("IsEmergencyNumber(%q, ZW) = true, want false", tt.number)
			}
		})
	}
}

// Upstream: ShortNumberInfoTest.testEmergencyNumberForSharedCountryCallingCode
func TestEmergencyNumberForSharedCountryCallingCode(t *testing.T) {
	info := Instance()

	// 112 is valid in both Australia and the Christmas Islands (shared CC 61).
	if !info.IsEmergencyNumber("112", "AU") {
		t.Error("IsEmergencyNumber(112, AU) = false, want true")
	}
	if !info.IsValidShortNumberForRegion(pn(61, 112), "AU") {
		t.Error("IsValidShortNumberForRegion(112, AU) = false, want true")
	}
	if got := info.GetExpectedCostForRegion(pn(61, 112), "AU"); got != phonesafe.CostTollFree {
		t.Errorf("GetExpectedCostForRegion(112, AU) = %v, want TollFree", got)
	}

	if !info.IsEmergencyNumber("112", "CX") {
		t.Error("IsEmergencyNumber(112, CX) = false, want true")
	}
	if !info.IsValidShortNumberForRegion(pn(61, 112), "CX") {
		t.Error("IsValidShortNumberForRegion(112, CX) = false, want true")
	}
	if got := info.GetExpectedCostForRegion(pn(61, 112), "CX"); got != phonesafe.CostTollFree {
		t.Errorf("GetExpectedCostForRegion(112, CX) = %v, want TollFree", got)
	}

	sharedEmergency := pn(61, 112)
	if !info.IsValidShortNumber(sharedEmergency) {
		t.Error("IsValidShortNumber(CC61 112) = false, want true")
	}
	if got := info.GetExpectedCost(sharedEmergency); got != phonesafe.CostTollFree {
		t.Errorf("GetExpectedCost(CC61 112) = %v, want TollFree", got)
	}
}

// Upstream: ShortNumberInfoTest.testOverlappingNANPANumber
func TestOverlappingNANPANumber(t *testing.T) {
	info := Instance()

	// 211 is an emergency number in Barbados, while it is a toll-free information
	// line in Canada and the USA.
	if !info.IsEmergencyNumber("211", "BB") {
		t.Error("IsEmergencyNumber(211, BB) = false, want true")
	}
	if got := info.GetExpectedCostForRegion(pn(1, 211), "BB"); got != phonesafe.CostTollFree {
		t.Errorf("GetExpectedCostForRegion(211, BB) = %v, want TollFree", got)
	}

	if info.IsEmergencyNumber("211", "US") {
		t.Error("IsEmergencyNumber(211, US) = true, want false")
	}
	if got := info.GetExpectedCostForRegion(pn(1, 211), "US"); got != phonesafe.CostUnknown {
		t.Errorf("GetExpectedCostForRegion(211, US) = %v, want Unknown", got)
	}

	if info.IsEmergencyNumber("211", "CA") {
		t.Error("IsEmergencyNumber(211, CA) = true, want false")
	}
	if got := info.GetExpectedCostForRegion(pn(1, 211), "CA"); got != phonesafe.CostTollFree {
		t.Errorf("GetExpectedCostForRegion(211, CA) = %v, want TollFree", got)
	}
}

// Upstream: ShortNumberInfoTest.testCountryCallingCodeIsNotIgnored
func TestCountryCallingCodeIsNotIgnored(t *testing.T) {
	info := Instance()

	// +46 is the country calling code for Sweden (SE), and 40404 is a valid short
	// number in the US. The CC should not be ignored.
	seNumber := pn(46, 40404)

	if info.IsPossibleShortNumberForRegion(seNumber, "US") {
		t.Error("IsPossibleShortNumberForRegion(+46 40404, US) = true, want false")
	}
	if info.IsValidShortNumberForRegion(seNumber, "US") {
		t.Error("IsValidShortNumberForRegion(+46 40404, US) = true, want false")
	}
	if got := info.GetExpectedCostForRegion(seNumber, "US"); got != phonesafe.CostUnknown {
		t.Errorf("GetExpectedCostForRegion(+46 40404, US) = %v, want Unknown", got)
	}
}

// Upstream: ShortNumberInfoTest.testGetExpectedCost (additional cases)
func TestGetExpectedCostInvalidButHasCost(t *testing.T) {
	info := Instance()

	// An invalid number may nevertheless have a cost other than UNKNOWN_COST.
	// FR 116123 is invalid but toll-free.
	frInvalid := pn(33, 116123)
	if info.IsValidShortNumberForRegion(frInvalid, "FR") {
		t.Error("IsValidShortNumberForRegion(116123, FR) = true, want false")
	}
	if got := info.GetExpectedCostForRegion(frInvalid, "FR"); got != phonesafe.CostTollFree {
		t.Errorf("GetExpectedCostForRegion(116123, FR) = %v, want TollFree", got)
	}
	if info.IsValidShortNumber(frInvalid) {
		t.Error("IsValidShortNumber(CC33 116123) = true, want false")
	}
	if got := info.GetExpectedCost(frInvalid); got != phonesafe.CostTollFree {
		t.Errorf("GetExpectedCost(CC33 116123) = %v, want TollFree", got)
	}

	// FR 12345 is unknown cost.
	frUnknown := pn(33, 12345)
	if got := info.GetExpectedCostForRegion(frUnknown, "FR"); got != phonesafe.CostUnknown {
		t.Errorf("GetExpectedCostForRegion(12345, FR) = %v, want Unknown", got)
	}
	if got := info.GetExpectedCost(frUnknown); got != phonesafe.CostUnknown {
		t.Errorf("GetExpectedCost(CC33 12345) = %v, want Unknown", got)
	}
}

// Upstream: ShortNumberInfoTest.testIsCarrierSpecific
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

// Upstream: ShortNumberInfoTest.testIsCarrierSpecific
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

// Upstream: ShortNumberInfoTest.testIsSmsService
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

// Upstream: ShortNumberInfoTest.testExampleShortNumberPresence
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
