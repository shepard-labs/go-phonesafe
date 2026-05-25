package phonesafe

import "github.com/shepard-labs/go-phonesafe/internal/metadata"

// GetNumberType returns the classification of a valid phone number.
// Returns TypeUnknown for invalid numbers.
func (u *PhoneNumberUtil) GetNumberType(number PhoneNumber) PhoneNumberType {
	region := u.getRegionCodeForNumber(number)
	meta := u.getMetadataForRegionOrCallingCode(number.CountryCode, region)
	if meta == nil {
		return TypeUnknown
	}
	nationalNumber := GetNationalSignificantNumber(number)
	return u.getNumberTypeHelper(nationalNumber, meta)
}

// getNumberTypeHelper determines the PhoneNumberType by testing the national
// number against each type descriptor in the metadata.
func (u *PhoneNumberUtil) getNumberTypeHelper(nationalNumber string, meta *metadata.PhoneMetadata) PhoneNumberType {
	// First check general desc — if it doesn't match, number is unknown.
	if !u.isNumberMatchingDesc(nationalNumber, meta.GeneralDesc) {
		return TypeUnknown
	}

	if u.isNumberMatchingDesc(nationalNumber, meta.PremiumRate) {
		return TypePremiumRate
	}
	if u.isNumberMatchingDesc(nationalNumber, meta.TollFree) {
		return TypeTollFree
	}
	if u.isNumberMatchingDesc(nationalNumber, meta.SharedCost) {
		return TypeSharedCost
	}
	if u.isNumberMatchingDesc(nationalNumber, meta.VOIP) {
		return TypeVOIP
	}
	if u.isNumberMatchingDesc(nationalNumber, meta.PersonalNumber) {
		return TypePersonalNumber
	}
	if u.isNumberMatchingDesc(nationalNumber, meta.Pager) {
		return TypePager
	}
	if u.isNumberMatchingDesc(nationalNumber, meta.UAN) {
		return TypeUAN
	}
	if u.isNumberMatchingDesc(nationalNumber, meta.Voicemail) {
		return TypeVoicemail
	}

	// Check fixed-line and mobile.
	isFixedLine := u.isNumberMatchingDesc(nationalNumber, meta.FixedLine)
	if isFixedLine {
		// If patterns are identical, it's FixedLineOrMobile.
		if meta.FixedLine.NationalNumberPattern == meta.Mobile.NationalNumberPattern {
			return TypeFixedLineOrMobile
		}
		if u.isNumberMatchingDesc(nationalNumber, meta.Mobile) {
			return TypeFixedLineOrMobile
		}
		return TypeFixedLine
	}

	// Check mobile only if patterns differ from fixed-line.
	if meta.FixedLine.NationalNumberPattern != meta.Mobile.NationalNumberPattern {
		if u.isNumberMatchingDesc(nationalNumber, meta.Mobile) {
			return TypeMobile
		}
	}

	return TypeUnknown
}

// IsNumberGeographical returns true if the number is associated with a
// geographic area (has an area code).
func (u *PhoneNumberUtil) IsNumberGeographical(number PhoneNumber) bool {
	numType := u.GetNumberType(number)
	return u.isNumberTypeGeographical(numType, number.CountryCode)
}

// isNumberTypeGeographical checks if a phone number type is geographical
// for the given country code.
func (u *PhoneNumberUtil) isNumberTypeGeographical(numType PhoneNumberType, cc int32) bool {
	return numType == TypeFixedLine || numType == TypeFixedLineOrMobile ||
		(numType == TypeMobile && isGeoMobileCountry(cc))
}

// geoMobileCountries lists country codes where mobile numbers are geographical.
// Matches upstream GEO_MOBILE_COUNTRIES set.
var geoMobileCountries = map[int32]bool{
	52: true, // Mexico
	54: true, // Argentina
	55: true, // Brazil
	62: true, // Indonesia
}

// isGeoMobileCountry returns true if the given country code has geographic mobile numbers.
func isGeoMobileCountry(cc int32) bool {
	return geoMobileCountries[cc]
}
