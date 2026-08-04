# Changelog

## v1.0.0

Initial release. Complete Go port of Google's libphonenumber covering:

### Core
- Parse phone numbers from international, national, vanity, and RFC3966 formats
- Format in E.164, international, national, and RFC3966 modes
- Full validation (IsPossible + IsValid) with detailed reason codes
- Number type classification (12 types including mobile, fixed-line, VOIP, toll-free)
- Number comparison with fuzzy matching (IsNumberMatch)
- As-you-type incremental formatting
- Phone number extraction from text (FindNumbers) with 4 leniency levels
- Utility functions: GetExampleNumber, TruncateTooLongNumber, GetSupportedRegions, etc.

### Short Numbers (`shortnumber` package)
- Short number validation and possibility checks
- Emergency number detection (exact and prefix matching)
- Cost classification (toll-free, standard rate, premium rate)
- Carrier-specific and SMS service detection

### Geocoder (`geocoder` package)
- Offline geographic descriptions for phone numbers
- 34 language support with English fallback
- User-region-aware country name omission

### Timezone (`timezone` package)
- Phone number prefix to IANA timezone mapping
- Country-level fallback for non-geographical numbers

### Metadata
- 240+ regions supported
- Code-generated from upstream libphonenumber XML resources
- Deterministic output via `cmd/phonesafe-gen`
