// Package shortnumber provides methods for short number validation and classification,
// including emergency number detection and cost categorization.
package shortnumber

import (
	"regexp"
	"slices"
	"strings"
	"sync"

	phonesafe "github.com/shepard-labs/go-phonesafe"
	"github.com/shepard-labs/go-phonesafe/internal/metadata"
)

// regionsWhereEmergencyNumbersMustBeExact lists countries where appending
// extra digits to an emergency number disconnects from emergency services.
var regionsWhereEmergencyNumbersMustBeExact = map[string]bool{
	"BR": true,
	"CL": true,
	"NI": true,
}

// plusCharsPattern matches a leading plus sign (ASCII or fullwidth).
var plusCharsPattern = regexp.MustCompile(`^[+\x{FF0B}]`)

// Info provides methods for short number validation and classification.
type Info struct {
	cache sync.Map // pattern string → *regexp.Regexp
}

var (
	instance *Info
	once     sync.Once
)

// Instance returns the global goroutine-safe ShortNumberInfo singleton.
func Instance() *Info {
	once.Do(func() {
		instance = &Info{}
	})
	return instance
}

// --- Private helpers ---

// getShortNumberMetadataForRegion returns the short number metadata for a region,
// or nil if none exists.
func getShortNumberMetadataForRegion(region string) *metadata.ShortPhoneMetadata {
	if region == "" {
		return nil
	}
	return metadata.ShortMetadata[region]
}

// regionDialingFromMatchesNumber checks that the number's country code corresponds
// to the given dialing region.
func regionDialingFromMatchesNumber(number phonesafe.PhoneNumber, region string) bool {
	regionCodes := metadata.CountryCodeToRegions[number.CountryCode]
	return slices.Contains(regionCodes, region)
}

// getRegionCodesForCountryCode returns all regions sharing a country calling code.
func getRegionCodesForCountryCode(cc int32) []string {
	return metadata.CountryCodeToRegions[cc]
}

// matchesEntirely returns true if the full string matches the regex pattern.
func (i *Info) matchesEntirely(pattern, number string) bool {
	if pattern == "" {
		return false
	}
	re := i.getRegexp(`^(?:` + pattern + `)$`)
	if re == nil {
		return false
	}
	return re.MatchString(number)
}

// matchesPrefix returns true if the beginning of number matches the pattern.
func (i *Info) matchesPrefix(pattern, number string) bool {
	if pattern == "" {
		return false
	}
	re := i.getRegexp(`^(?:` + pattern + `)`)
	if re == nil {
		return false
	}
	return re.MatchString(number)
}

// getRegexp compiles and caches a regexp pattern.
func (i *Info) getRegexp(pattern string) *regexp.Regexp {
	if v, ok := i.cache.Load(pattern); ok {
		return v.(*regexp.Regexp)
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	i.cache.Store(pattern, re)
	return re
}

// matchesPossibleNumberAndNationalNumber checks both length and pattern match.
func (i *Info) matchesPossibleNumberAndNationalNumber(number string, desc metadata.PhoneNumberDesc) bool {
	if len(desc.PossibleLengths) > 0 && !slices.Contains(desc.PossibleLengths, int32(len(number))) {
		return false
	}
	return i.matchesEntirely(desc.NationalNumberPattern, number)
}

// getRegionCodeForShortNumberFromRegionList finds the first region where the number
// is a valid short code among the given regions.
func (i *Info) getRegionCodeForShortNumberFromRegionList(number phonesafe.PhoneNumber, regionCodes []string) string {
	if len(regionCodes) == 0 {
		return ""
	}
	if len(regionCodes) == 1 {
		return regionCodes[0]
	}
	nationalNumber := phonesafe.GetNationalSignificantNumber(number)
	for _, region := range regionCodes {
		meta := getShortNumberMetadataForRegion(region)
		if meta != nil && i.matchesPossibleNumberAndNationalNumber(nationalNumber, meta.GeneralDesc) {
			return region
		}
	}
	return ""
}

// extractPossibleNumber strips formatting to extract a possible phone number.
// Simplified version — finds first digit/plus and strips trailing junk.
func extractPossibleNumber(number string) string {
	// Find first valid start char (digit or plus).
	idx := strings.IndexAny(number, "+＋0123456789٠١٢٣٤٥٦٧٨٩۰۱۲۳۴۵۶۷۸۹")
	if idx < 0 {
		return ""
	}
	return number[idx:]
}

// --- Public methods ---

// IsPossibleShortNumber checks if a short number is possible for any region
// matching its country code.
func (i *Info) IsPossibleShortNumber(number phonesafe.PhoneNumber) bool {
	regionCodes := getRegionCodesForCountryCode(number.CountryCode)
	shortNumberLength := int32(len(phonesafe.GetNationalSignificantNumber(number)))
	for _, region := range regionCodes {
		meta := getShortNumberMetadataForRegion(region)
		if meta == nil {
			continue
		}
		if slices.Contains(meta.GeneralDesc.PossibleLengths, shortNumberLength) {
			return true
		}
	}
	return false
}

// IsPossibleShortNumberForRegion checks possibility for a specific region.
func (i *Info) IsPossibleShortNumberForRegion(number phonesafe.PhoneNumber, region string) bool {
	if !regionDialingFromMatchesNumber(number, region) {
		return false
	}
	meta := getShortNumberMetadataForRegion(region)
	if meta == nil {
		return false
	}
	numberLength := int32(len(phonesafe.GetNationalSignificantNumber(number)))
	return slices.Contains(meta.GeneralDesc.PossibleLengths, numberLength)
}

// IsValidShortNumber checks if a short number matches a valid pattern.
func (i *Info) IsValidShortNumber(number phonesafe.PhoneNumber) bool {
	regionCodes := getRegionCodesForCountryCode(number.CountryCode)
	regionCode := i.getRegionCodeForShortNumberFromRegionList(number, regionCodes)
	if len(regionCodes) > 1 && regionCode != "" {
		// If a matching region was found among multiple regions, validity is implied.
		return true
	}
	return i.IsValidShortNumberForRegion(number, regionCode)
}

// IsValidShortNumberForRegion validates for a specific dialing region.
func (i *Info) IsValidShortNumberForRegion(number phonesafe.PhoneNumber, region string) bool {
	if !regionDialingFromMatchesNumber(number, region) {
		return false
	}
	meta := getShortNumberMetadataForRegion(region)
	if meta == nil {
		return false
	}
	shortNumber := phonesafe.GetNationalSignificantNumber(number)
	if !i.matchesPossibleNumberAndNationalNumber(shortNumber, meta.GeneralDesc) {
		return false
	}
	// GeneralDesc serves as the combined valid short code pattern.
	return i.matchesPossibleNumberAndNationalNumber(shortNumber, meta.GeneralDesc)
}

// GetExpectedCost returns the cost category of a short number.
// For shared country codes, returns the highest cost across all regions.
func (i *Info) GetExpectedCost(number phonesafe.PhoneNumber) phonesafe.ShortNumberCost {
	regionCodes := getRegionCodesForCountryCode(number.CountryCode)
	if len(regionCodes) == 0 {
		return phonesafe.CostUnknown
	}
	if len(regionCodes) == 1 {
		return i.GetExpectedCostForRegion(number, regionCodes[0])
	}
	// Multiple regions: priority is PREMIUM > UNKNOWN > STANDARD > TOLL_FREE
	cost := phonesafe.CostTollFree
	for _, region := range regionCodes {
		costForRegion := i.GetExpectedCostForRegion(number, region)
		switch costForRegion {
		case phonesafe.CostPremiumRate:
			return phonesafe.CostPremiumRate
		case phonesafe.CostUnknown:
			cost = phonesafe.CostUnknown
		case phonesafe.CostStandardRate:
			if cost != phonesafe.CostUnknown {
				cost = phonesafe.CostStandardRate
			}
		case phonesafe.CostTollFree:
			// No escalation.
		}
	}
	return cost
}

// GetExpectedCostForRegion returns cost when dialing from a specific region.
func (i *Info) GetExpectedCostForRegion(number phonesafe.PhoneNumber, region string) phonesafe.ShortNumberCost {
	if !regionDialingFromMatchesNumber(number, region) {
		return phonesafe.CostUnknown
	}
	meta := getShortNumberMetadataForRegion(region)
	if meta == nil {
		return phonesafe.CostUnknown
	}

	shortNumber := phonesafe.GetNationalSignificantNumber(number)

	// Early exit if length doesn't match general desc.
	if !slices.Contains(meta.GeneralDesc.PossibleLengths, int32(len(shortNumber))) {
		return phonesafe.CostUnknown
	}

	// Test in order of decreasing expense.
	if i.matchesPossibleNumberAndNationalNumber(shortNumber, meta.PremiumRate) {
		return phonesafe.CostPremiumRate
	}
	if i.matchesPossibleNumberAndNationalNumber(shortNumber, meta.StandardRate) {
		return phonesafe.CostStandardRate
	}
	if i.matchesPossibleNumberAndNationalNumber(shortNumber, meta.TollFree) {
		return phonesafe.CostTollFree
	}
	if i.IsEmergencyNumber(shortNumber, region) {
		// Emergency numbers are implicitly toll-free.
		return phonesafe.CostTollFree
	}
	return phonesafe.CostUnknown
}

// ConnectsToEmergencyNumber returns true if the number might connect to emergency services.
// Allows extra appended digits in certain countries.
func (i *Info) ConnectsToEmergencyNumber(number, region string) bool {
	return i.matchesEmergencyNumberHelper(number, region, true)
}

// IsEmergencyNumber returns true if the number exactly matches an emergency number.
func (i *Info) IsEmergencyNumber(number, region string) bool {
	return i.matchesEmergencyNumberHelper(number, region, false)
}

// matchesEmergencyNumberHelper implements emergency number matching with optional prefix matching.
func (i *Info) matchesEmergencyNumberHelper(number, region string, allowPrefixMatch bool) bool {
	possibleNumber := extractPossibleNumber(number)
	if possibleNumber == "" {
		return false
	}

	// Reject numbers starting with plus sign.
	if plusCharsPattern.MatchString(possibleNumber) {
		return false
	}

	meta := getShortNumberMetadataForRegion(region)
	if meta == nil || meta.Emergency.NationalNumberPattern == "" {
		return false
	}

	normalizedNumber := phonesafe.NormalizeDigitsOnly(possibleNumber)
	if normalizedNumber == "" {
		return false
	}

	allowPrefixMatchForRegion := allowPrefixMatch && !regionsWhereEmergencyNumbersMustBeExact[region]
	if allowPrefixMatchForRegion {
		return i.matchesPrefix(meta.Emergency.NationalNumberPattern, normalizedNumber)
	}
	return i.matchesEntirely(meta.Emergency.NationalNumberPattern, normalizedNumber)
}

// IsCarrierSpecific returns true if the short number is carrier-specific.
func (i *Info) IsCarrierSpecific(number phonesafe.PhoneNumber) bool {
	regionCodes := getRegionCodesForCountryCode(number.CountryCode)
	regionCode := i.getRegionCodeForShortNumberFromRegionList(number, regionCodes)
	nationalNumber := phonesafe.GetNationalSignificantNumber(number)
	meta := getShortNumberMetadataForRegion(regionCode)
	return meta != nil && i.matchesPossibleNumberAndNationalNumber(nationalNumber, meta.CarrierSpecific)
}

// IsCarrierSpecificForRegion checks carrier-specificity for a dialing region.
func (i *Info) IsCarrierSpecificForRegion(number phonesafe.PhoneNumber, region string) bool {
	if !regionDialingFromMatchesNumber(number, region) {
		return false
	}
	nationalNumber := phonesafe.GetNationalSignificantNumber(number)
	meta := getShortNumberMetadataForRegion(region)
	return meta != nil && i.matchesPossibleNumberAndNationalNumber(nationalNumber, meta.CarrierSpecific)
}

// IsSmsServiceForRegion returns true if the number is an SMS service for the region.
func (i *Info) IsSmsServiceForRegion(number phonesafe.PhoneNumber, region string) bool {
	if !regionDialingFromMatchesNumber(number, region) {
		return false
	}
	meta := getShortNumberMetadataForRegion(region)
	return meta != nil && i.matchesPossibleNumberAndNationalNumber(
		phonesafe.GetNationalSignificantNumber(number), meta.SMSServices)
}

// GetExampleShortNumber returns an example short code for the region.
func (i *Info) GetExampleShortNumber(region string) string {
	meta := getShortNumberMetadataForRegion(region)
	if meta == nil {
		return ""
	}
	if meta.GeneralDesc.ExampleNumber != "" {
		return meta.GeneralDesc.ExampleNumber
	}
	return ""
}

// GetExampleShortNumberForCost returns an example of the given cost category.
func (i *Info) GetExampleShortNumberForCost(region string, cost phonesafe.ShortNumberCost) string {
	meta := getShortNumberMetadataForRegion(region)
	if meta == nil {
		return ""
	}
	var desc metadata.PhoneNumberDesc
	switch cost {
	case phonesafe.CostTollFree:
		desc = meta.TollFree
	case phonesafe.CostStandardRate:
		desc = meta.StandardRate
	case phonesafe.CostPremiumRate:
		desc = meta.PremiumRate
	default:
		return ""
	}
	if desc.ExampleNumber != "" {
		return desc.ExampleNumber
	}
	return ""
}
