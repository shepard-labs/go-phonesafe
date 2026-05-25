// Package data holds generated geocoding and timezone lookup data.
package data

// GeoEntry maps a phone number prefix to a geographic description.
type GeoEntry struct {
	Prefix      int32
	Description string
}

// TZEntry maps a phone number prefix to a list of IANA timezone IDs.
type TZEntry struct {
	Prefix    int32
	TimeZones []string
}
