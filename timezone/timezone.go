// Package timezone provides offline phone number to IANA timezone mapping.
// It maps phone number prefixes to timezone identifiers using generated data
// from libphonenumber's timezone mapping files.
package timezone

import (
	"fmt"
	"sort"
	"strconv"
	"sync"

	phonesafe "github.com/shepard-labs/go-phonesafe"
	"github.com/shepard-labs/go-phonesafe/internal/data"
)

// UnknownTimeZone is the fallback timezone ID returned when no mapping is found.
const UnknownTimeZone = "Etc/Unknown"

// unknownTimeZoneList is a pre-allocated slice for the unknown case.
var unknownTimeZoneList = []string{UnknownTimeZone}

// Mapper provides phone number to IANA timezone mapping.
type Mapper struct{}

var (
	instance *Mapper
	once     sync.Once
)

// Instance returns the global goroutine-safe timezone Mapper singleton.
func Instance() *Mapper {
	once.Do(func() {
		instance = &Mapper{}
	})
	return instance
}

// GetTimeZonesForNumber returns IANA timezone IDs for the given number.
// Returns ["Etc/Unknown"] if the number is invalid or unmappable.
func (m *Mapper) GetTimeZonesForNumber(number phonesafe.PhoneNumber) []string {
	util := phonesafe.Instance()
	numType := util.GetNumberType(number)
	if numType == phonesafe.TypeUnknown {
		return unknownTimeZoneList
	}
	if !util.IsNumberGeographical(number) {
		return m.getCountryLevelTimeZones(number)
	}
	return m.GetTimeZonesForGeographicalNumber(number)
}

// GetTimeZonesForGeographicalNumber returns timezones for a number known to be geographical.
// Caller should verify IsNumberGeographical first for best results.
// Falls back to country-level timezones if no specific prefix match is found.
// Returns ["Etc/Unknown"] if no timezones can be determined.
func (m *Mapper) GetTimeZonesForGeographicalNumber(number phonesafe.PhoneNumber) []string {
	timezones := m.lookupTimezones(number)
	if len(timezones) > 0 {
		return timezones
	}
	return unknownTimeZoneList
}

// lookupTimezones performs a longest-prefix match against timezone data.
// Builds the full E.164 prefix (CC + national number) and progressively
// strips trailing digits until a match is found.
func (m *Mapper) lookupTimezones(number phonesafe.PhoneNumber) []string {
	nsn := phonesafe.GetNationalSignificantNumber(number)
	prefix := fmt.Sprintf("%d%s", number.CountryCode, nsn)

	entries := data.TimezoneData

	// Progressively strip trailing digits to find the longest matching prefix.
	for len(prefix) > 0 {
		prefixInt, err := strconv.ParseInt(prefix, 10, 32)
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
			return entries[idx].TimeZones
		}
		prefix = prefix[:len(prefix)-1]
	}
	return nil
}

// getCountryLevelTimeZones returns timezones for the number's country code.
// Used for non-geographical numbers that can't be pinpointed to a specific area.
func (m *Mapper) getCountryLevelTimeZones(number phonesafe.PhoneNumber) []string {
	entries := data.TimezoneData
	cc := number.CountryCode

	// Binary search for the country code prefix.
	idx := sort.Search(len(entries), func(i int) bool {
		return entries[i].Prefix >= cc
	})
	if idx < len(entries) && entries[idx].Prefix == cc {
		return entries[idx].TimeZones
	}
	return unknownTimeZoneList
}
