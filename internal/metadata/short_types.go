package metadata

// ShortPhoneMetadata holds short number metadata for a single region.
type ShortPhoneMetadata struct {
	GeneralDesc       PhoneNumberDesc
	TollFree          PhoneNumberDesc
	StandardRate      PhoneNumberDesc
	PremiumRate       PhoneNumberDesc
	CarrierSpecific   PhoneNumberDesc
	Emergency         PhoneNumberDesc
	ExpandedEmergency PhoneNumberDesc
	SMSServices       PhoneNumberDesc
}
