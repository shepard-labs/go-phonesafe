// Package geocoder provides offline geographic descriptions for phone numbers.
// It maps phone number prefixes to location names (e.g., "Mountain View, CA", "Seoul").
package geocoder

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/text/language"
	"golang.org/x/text/language/display"

	phonesafe "github.com/shepard-labs/go-phonesafe"
	"github.com/shepard-labs/go-phonesafe/internal/data"
)

// Geocoder provides offline geographic descriptions for phone numbers.
type Geocoder struct{}

var (
	instance *Geocoder
	once     sync.Once
)

// Instance returns the global goroutine-safe Geocoder singleton.
func Instance() *Geocoder {
	once.Do(func() {
		instance = &Geocoder{}
	})
	return instance
}

// GetDescriptionForNumber returns a geographic description for the given phone number
// in the specified language. Validates the number before geocoding.
// Returns empty string if the number is invalid or unlocatable.
func (g *Geocoder) GetDescriptionForNumber(number phonesafe.PhoneNumber, lang language.Tag) string {
	util := phonesafe.Instance()
	numType := util.GetNumberType(number)
	if numType == phonesafe.TypeUnknown {
		return ""
	}
	if !util.IsNumberGeographical(number) {
		return g.getCountryNameForNumber(number, lang)
	}
	return g.GetDescriptionForValidNumber(number, lang)
}

// GetDescriptionForValidNumber returns a geographic description for a number
// assumed to be valid. Caller must ensure the number is valid.
// Returns the area description if available, otherwise the country name.
func (g *Geocoder) GetDescriptionForValidNumber(number phonesafe.PhoneNumber, lang language.Tag) string {
	langKey := langTagToKey(lang)

	var areaDescription string
	mobileToken := phonesafe.GetCountryMobileToken(number.CountryCode)
	nationalNumber := phonesafe.GetNationalSignificantNumber(number)

	if mobileToken != "" && strings.HasPrefix(nationalNumber, mobileToken) {
		// In some countries (e.g., Argentina), mobile numbers have a mobile token
		// before the national destination code. Strip it before geocoding.
		strippedNational := nationalNumber[len(mobileToken):]
		// Re-parse to get a number suitable for prefix lookup.
		util := phonesafe.Instance()
		region := util.GetRegionCodeForNumber(number)
		copiedNumber, err := util.Parse(strippedNational, region)
		if err != nil {
			copiedNumber = number
		}
		areaDescription = g.lookupDescription(copiedNumber, langKey)
	} else {
		areaDescription = g.lookupDescription(number, langKey)
	}

	if areaDescription != "" {
		return areaDescription
	}
	return g.getCountryNameForNumber(number, lang)
}

// GetDescriptionForValidNumberWithUserRegion omits the country name if the number
// is from the same country as the user.
// Example: US user sees "Mountain View, CA" instead of "Mountain View, CA, United States".
func (g *Geocoder) GetDescriptionForValidNumberWithUserRegion(number phonesafe.PhoneNumber, lang language.Tag, userRegion string) string {
	util := phonesafe.Instance()
	regionCode := util.GetRegionCodeForNumber(number)
	if strings.EqualFold(userRegion, regionCode) {
		return g.GetDescriptionForValidNumber(number, lang)
	}
	// User is in a different region — show just the country/region name.
	return g.getRegionDisplayName(regionCode, lang)
}

// lookupDescription performs a longest-prefix match against geocoding data.
// Falls back to English if no data exists for the requested language.
func (g *Geocoder) lookupDescription(number phonesafe.PhoneNumber, langKey string) string {
	desc := g.lookupInLanguage(number, langKey)
	if desc != "" {
		return desc
	}
	// Fall back to English unless the language is CJK (they have their own data
	// and shouldn't show English fallback).
	if langKey != "en" && mayFallBackToEnglish(langKey) {
		return g.lookupInLanguage(number, "en")
	}
	return ""
}

// lookupInLanguage performs a longest-prefix match in the given language's data.
func (g *Geocoder) lookupInLanguage(number phonesafe.PhoneNumber, langKey string) string {
	entries := data.GeoData[langKey]
	if len(entries) == 0 {
		return ""
	}

	// Build the full E.164 number as an integer (country code + national number).
	nsn := phonesafe.GetNationalSignificantNumber(number)
	prefix := fmt.Sprintf("%d%s", number.CountryCode, nsn)

	// Progressively strip trailing digits to find the longest matching prefix.
	for len(prefix) > 0 {
		prefixInt, err := strconv.ParseInt(prefix, 10, 64)
		if err != nil {
			prefix = prefix[:len(prefix)-1]
			continue
		}
		prefixI32 := int32(prefixInt)
		// Binary search for this prefix.
		idx := sort.Search(len(entries), func(i int) bool {
			return entries[i].Prefix >= prefixI32
		})
		if idx < len(entries) && entries[idx].Prefix == prefixI32 {
			return entries[idx].Description
		}
		prefix = prefix[:len(prefix)-1]
	}
	return ""
}

// getCountryNameForNumber returns the country display name for a phone number.
// Returns empty string if the number belongs to multiple countries ambiguously.
func (g *Geocoder) getCountryNameForNumber(number phonesafe.PhoneNumber, lang language.Tag) string {
	util := phonesafe.Instance()
	regionCodes := util.GetRegionCodesForCountryCode(number.CountryCode)
	if len(regionCodes) == 1 {
		return g.getRegionDisplayName(regionCodes[0], lang)
	}
	if len(regionCodes) > 1 {
		// Multiple regions share this calling code — disambiguate.
		regionWhereValid := "ZZ"
		for _, region := range regionCodes {
			if util.IsValidNumberForRegion(number, region) {
				if regionWhereValid != "ZZ" {
					// Valid for multiple regions — ambiguous.
					return ""
				}
				regionWhereValid = region
			}
		}
		return g.getRegionDisplayName(regionWhereValid, lang)
	}
	return ""
}

// getRegionDisplayName returns the localized display name of a region.
// Returns empty string for "ZZ" or "001" (non-geographical).
func (g *Geocoder) getRegionDisplayName(regionCode string, lang language.Tag) string {
	if regionCode == "" || regionCode == "ZZ" || regionCode == "001" {
		return ""
	}
	region, err := language.ParseRegion(regionCode)
	if err != nil {
		return ""
	}
	return display.Regions(lang).Name(region)
}

// langTagToKey extracts a geocoding data lookup key from a language.Tag.
// Handles script subtags for zh_Hant.
func langTagToKey(lang language.Tag) string {
	base, script, _ := lang.Raw()
	langStr := base.String()

	// Hebrew canonical form in Java is "iw" (not "he").
	if langStr == "he" {
		langStr = "iw"
	}

	// Check if lang+script combination exists (e.g., "zh_Hant").
	var zeroScript language.Script
	if script != zeroScript {
		withScript := langStr + "_" + script.String()
		if _, ok := data.GeoData[withScript]; ok {
			return withScript
		}
	}

	// Check base language.
	if _, ok := data.GeoData[langStr]; ok {
		return langStr
	}
	return ""
}

// mayFallBackToEnglish returns false for CJK languages which should not
// fall back to English descriptions.
func mayFallBackToEnglish(langKey string) bool {
	return langKey != "zh" && langKey != "zh_Hant" && langKey != "ja" && langKey != "ko"
}
