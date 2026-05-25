// Package phonesafe provides parsing, formatting, and validation of international
// phone numbers. It is a Go port of Google's libphonenumber.
package phonesafe

import "fmt"

// PhoneNumber represents a parsed international telephone number.
type PhoneNumber struct {
	// CountryCode is the ITU country calling code (e.g., 1 for NANPA, 44 for UK).
	CountryCode int32

	// NationalNumber is the national significant number without any prefix.
	NationalNumber uint64

	// Extension is the phone number extension (digits, possibly with wait chars like ',').
	Extension string

	// ItalianLeadingZero indicates the national number has a significant leading zero.
	ItalianLeadingZero bool

	// NumberOfLeadingZeros is the count of significant leading zeros (default 1 when ItalianLeadingZero is true).
	NumberOfLeadingZeros int32

	// RawInput is the original input string before parsing. Only populated when
	// ParseAndKeepRawInput is used.
	RawInput string

	// CountryCodeSource indicates how the country code was derived during parsing.
	// Only populated when ParseAndKeepRawInput is used.
	CountryCodeSource CountryCodeSource

	// PreferredDomesticCarrierCode is the carrier selection code for domestic dialing.
	PreferredDomesticCarrierCode string
}

// CountryCodeSource indicates how the country code was derived during parsing.
type CountryCodeSource int

const (
	CountryCodeUnspecified           CountryCodeSource = 0
	CountryCodeFromNumberWithPlus    CountryCodeSource = 1
	CountryCodeFromNumberWithIDD     CountryCodeSource = 5
	CountryCodeFromNumberWithoutPlus CountryCodeSource = 10
	CountryCodeFromDefaultCountry    CountryCodeSource = 20
)

// PhoneNumberFormat specifies how to format a phone number.
type PhoneNumberFormat int

const (
	FormatE164          PhoneNumberFormat = iota // "+41446681800"
	FormatInternational                         // "+41 44 668 1800"
	FormatNational                              // "044 668 1800"
	FormatRFC3966                               // "tel:+41-44-668-1800"
)

// PhoneNumberType classifies a phone number.
type PhoneNumberType int

const (
	TypeFixedLine PhoneNumberType = iota
	TypeMobile
	TypeFixedLineOrMobile
	TypeTollFree
	TypePremiumRate
	TypeSharedCost
	TypeVOIP
	TypePersonalNumber
	TypePager
	TypeUAN
	TypeVoicemail
	TypeUnknown
)

// MatchType indicates the confidence level of a number comparison.
type MatchType int

const (
	MatchInvalidNumber MatchType = iota
	MatchNone
	MatchShortNSN
	MatchNSN
	MatchExact
)

// ValidationResult indicates whether a number is possible.
type ValidationResult int

const (
	IsPossible          ValidationResult = iota
	IsPossibleLocalOnly
	InvalidCountryCode
	TooShort
	InvalidLength
	TooLong
)

// ShortNumberCost indicates the cost category of a short number.
type ShortNumberCost int

const (
	CostTollFree ShortNumberCost = iota
	CostStandardRate
	CostPremiumRate
	CostUnknown
)

// Leniency controls how strictly PhoneNumberMatcher validates candidates.
type Leniency int

const (
	LeniencyPossible Leniency = iota
	LeniencyValid
	LeniencyStrictGrouping
	LeniencyExactGrouping
)

// ErrorCode represents the specific reason a phone number could not be parsed.
type ErrorCode int

const (
	ErrNone ErrorCode = iota
	ErrInvalidCountryCode
	ErrNotANumber
	ErrTooShortAfterIDD
	ErrTooShortNSN
	ErrTooLongNSN
)

// ParseError represents a phone number parsing failure.
type ParseError struct {
	Code    ErrorCode
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("phonesafe: %s", e.Message)
}

// NumberFormatRule defines a caller-supplied formatting pattern for FormatByPattern.
type NumberFormatRule struct {
	Pattern                      string
	Format                       string
	LeadingDigitsPatterns        []string
	NationalPrefixFormattingRule string
	NationalPrefixOptional       bool
	DomesticCarrierCodeFormattingRule string
}

// PhoneNumberMatch represents a phone number found in text.
type PhoneNumberMatch struct {
	// Start is the index into the searched text where the match begins.
	Start int

	// End is the exclusive end index of the match.
	End int

	// RawString is the matched substring from the searched text.
	RawString string

	// Number is the parsed phone number.
	Number PhoneNumber
}
