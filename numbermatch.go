package phonesafe

import (
	"strconv"
	"strings"
)

// copyCoreFieldsOnly returns a copy containing only the fields relevant for
// number matching: CountryCode, NationalNumber, Extension, ItalianLeadingZero,
// and NumberOfLeadingZeros. RawInput, CountryCodeSource, and
// PreferredDomesticCarrierCode are excluded because they don't define the
// number's identity.
func copyCoreFieldsOnly(n PhoneNumber) PhoneNumber {
	result := PhoneNumber{
		CountryCode:    n.CountryCode,
		NationalNumber: n.NationalNumber,
		Extension:      n.Extension,
	}
	if n.ItalianLeadingZero {
		result.ItalianLeadingZero = true
		result.NumberOfLeadingZeros = n.NumberOfLeadingZeros
	}
	return result
}

// exactlySameAs compares two PhoneNumber structs for exact equality on core
// fields. When ItalianLeadingZero is false, NumberOfLeadingZeros is ignored
// (it's only meaningful when leading zeros are significant).
func exactlySameAs(a, b PhoneNumber) bool {
	if a.CountryCode != b.CountryCode {
		return false
	}
	if a.NationalNumber != b.NationalNumber {
		return false
	}
	if a.Extension != b.Extension {
		return false
	}
	if a.ItalianLeadingZero != b.ItalianLeadingZero {
		return false
	}
	if a.ItalianLeadingZero {
		// Only compare leading zeros count when both have ItalianLeadingZero set.
		// Default is 1 when ItalianLeadingZero is true but NumberOfLeadingZeros is 0.
		aZeros := a.NumberOfLeadingZeros
		if aZeros == 0 {
			aZeros = 1
		}
		bZeros := b.NumberOfLeadingZeros
		if bZeros == 0 {
			bZeros = 1
		}
		if aZeros != bZeros {
			return false
		}
	}
	return true
}

// isNationalNumberSuffixOfTheOther returns true when one national number
// is the suffix of the other or both are the same.
func isNationalNumberSuffixOfTheOther(a, b PhoneNumber) bool {
	aNsn := strconv.FormatUint(a.NationalNumber, 10)
	bNsn := strconv.FormatUint(b.NationalNumber, 10)
	return strings.HasSuffix(aNsn, bNsn) || strings.HasSuffix(bNsn, aNsn)
}

// IsNumberMatch compares two phone numbers for equality with fuzzy matching.
// Returns one of: MatchExact, MatchNSN, MatchShortNSN, MatchNone, or
// MatchInvalidNumber (never returned from this overload — only from string variant).
func (u *PhoneNumberUtil) IsNumberMatch(firstNumberIn, secondNumberIn PhoneNumber) MatchType {
	// Copy core fields only — ignore raw input, source, carrier code.
	first := copyCoreFieldsOnly(firstNumberIn)
	second := copyCoreFieldsOnly(secondNumberIn)

	// Early exit if both have extensions and they differ.
	if first.Extension != "" && second.Extension != "" &&
		first.Extension != second.Extension {
		return MatchNone
	}

	firstCC := first.CountryCode
	secondCC := second.CountryCode

	// Both have country codes specified.
	if firstCC != 0 && secondCC != 0 {
		if exactlySameAs(first, second) {
			return MatchExact
		}
		if firstCC == secondCC && isNationalNumberSuffixOfTheOther(first, second) {
			return MatchShortNSN
		}
		return MatchNone
	}

	// One or both country codes are not specified.
	// Set them equal to simplify comparison.
	first.CountryCode = secondCC
	if exactlySameAs(first, second) {
		return MatchNSN
	}
	if isNationalNumberSuffixOfTheOther(first, second) {
		return MatchShortNSN
	}
	return MatchNone
}

// IsNumberMatchString parses both strings and compares them.
// Returns MatchInvalidNumber if either string is not a viable phone number.
func (u *PhoneNumberUtil) IsNumberMatchString(firstNumber, secondNumber string) MatchType {
	// Try parsing the first number with no default region.
	firstProto, err := u.Parse(firstNumber, "ZZ")
	if err == nil {
		return u.isNumberMatchOneString(firstProto, secondNumber)
	}

	// If the error was INVALID_COUNTRY_CODE, try parsing the second number.
	if pe, ok := err.(*ParseError); ok && pe.Code == ErrInvalidCountryCode {
		secondProto, err2 := u.Parse(secondNumber, "ZZ")
		if err2 == nil {
			return u.isNumberMatchOneString(secondProto, firstNumber)
		}
		if pe2, ok2 := err2.(*ParseError); ok2 && pe2.Code == ErrInvalidCountryCode {
			// Both have invalid country codes — try parsing without region checks.
			firstProto2, err3 := u.parseHelper(firstNumber, "", false, false)
			if err3 == nil {
				secondProto2, err4 := u.parseHelper(secondNumber, "", false, false)
				if err4 == nil {
					return u.IsNumberMatch(firstProto2, secondProto2)
				}
			}
		}
	}

	return MatchInvalidNumber
}

// isNumberMatchOneString compares a parsed PhoneNumber against a string.
func (u *PhoneNumberUtil) isNumberMatchOneString(firstNumber PhoneNumber, secondNumber string) MatchType {
	// Try parsing the second number with no default region.
	secondProto, err := u.Parse(secondNumber, "ZZ")
	if err == nil {
		return u.IsNumberMatch(firstNumber, secondProto)
	}

	if pe, ok := err.(*ParseError); ok && pe.Code == ErrInvalidCountryCode {
		// The second number has no country calling code. EXACT_MATCH is no longer
		// possible. Parse it using the first number's region.
		firstRegion := u.getRegionCodeForCountryCode(firstNumber.CountryCode)
		if firstRegion != "ZZ" {
			secondWithRegion, err2 := u.Parse(secondNumber, firstRegion)
			if err2 == nil {
				match := u.IsNumberMatch(firstNumber, secondWithRegion)
				if match == MatchExact {
					return MatchNSN
				}
				return match
			}
		}
		// If the first number didn't have a valid country calling code, parse
		// the second number without region checks.
		secondProto2, err3 := u.parseHelper(secondNumber, "", false, false)
		if err3 == nil {
			return u.IsNumberMatch(firstNumber, secondProto2)
		}
	}

	return MatchInvalidNumber
}
