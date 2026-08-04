// Package metadata holds generated phone number metadata types and data.
package metadata

// PhoneMetadata holds the numbering plan rules for a single region or non-geo entity.
type PhoneMetadata struct {
	ID                           string
	CountryCode                  int32
	InternationalPrefix          string
	PreferredInternationalPrefix string
	NationalPrefix               string
	PreferredExtnPrefix          string
	NationalPrefixForParsing     string
	NationalPrefixTransformRule  string
	MainCountryForCode           bool
	LeadingDigits                string
	MobileNumberPortable         bool

	GeneralDesc             PhoneNumberDesc
	FixedLine               PhoneNumberDesc
	Mobile                  PhoneNumberDesc
	TollFree                PhoneNumberDesc
	PremiumRate             PhoneNumberDesc
	SharedCost              PhoneNumberDesc
	PersonalNumber          PhoneNumberDesc
	VOIP                    PhoneNumberDesc
	Pager                   PhoneNumberDesc
	UAN                     PhoneNumberDesc
	Voicemail               PhoneNumberDesc
	NoInternationalDialling PhoneNumberDesc

	NumberFormats     []NumberFormat
	IntlNumberFormats []NumberFormat
}

// PhoneNumberDesc describes patterns and lengths for a specific phone number type.
type PhoneNumberDesc struct {
	NationalNumberPattern string
	PossibleLengths       []int32
	PossibleLengthsLocal  []int32
	ExampleNumber         string
}

// NumberFormat defines a formatting rule for phone numbers.
type NumberFormat struct {
	Pattern                           string
	Format                            string
	LeadingDigitsPatterns             []string
	NationalPrefixFormattingRule      string
	NationalPrefixOptional            bool
	DomesticCarrierCodeFormattingRule string
	IntlFormat                        string
}
