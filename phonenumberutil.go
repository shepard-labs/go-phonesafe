package phonesafe

import "sync"

// PhoneNumberUtil is the primary entry point for phone number operations.
// It is goroutine-safe and should be obtained via Instance().
type PhoneNumberUtil struct {
	rc *regexpCache
}

var (
	instance *PhoneNumberUtil
	once     sync.Once
)

// Instance returns the global goroutine-safe PhoneNumberUtil singleton.
// Initialized on first call via sync.Once.
func Instance() *PhoneNumberUtil {
	once.Do(func() {
		instance = &PhoneNumberUtil{
			rc: newRegexpCache(),
		}
	})
	return instance
}

// Parse parses a phone number string with a default region context.
// The number may contain formatting, extensions, and alpha characters.
// Returns ParseError if the input cannot be interpreted as a phone number.
//
// defaultRegion is a CLDR two-letter uppercase region code (e.g., "US", "GB").
// Use "ZZ" if the number is guaranteed to start with "+".
func (u *PhoneNumberUtil) Parse(numberToParse, defaultRegion string) (PhoneNumber, error) {
	return u.parseHelper(numberToParse, defaultRegion, false, true)
}

// ParseAndKeepRawInput is like Parse but also populates RawInput and
// CountryCodeSource fields for round-trip formatting.
func (u *PhoneNumberUtil) ParseAndKeepRawInput(numberToParse, defaultRegion string) (PhoneNumber, error) {
	return u.parseHelper(numberToParse, defaultRegion, true, true)
}
