# go-phonesafe

Go port of Google's [libphonenumber](https://github.com/google/libphonenumber). Parse, format, and validate international phone numbers.

[![Go Reference](https://pkg.go.dev/badge/github.com/shepard-labs/go-phonesafe.svg)](https://pkg.go.dev/github.com/shepard-labs/go-phonesafe)

## Features

- **Parse** phone numbers from any format (international, national, vanity, RFC3966)
- **Format** in E.164, international, national, or RFC3966 format
- **Validate** with length checks and full pattern matching
- **Classify** number types (mobile, fixed-line, toll-free, premium rate, etc.)
- **Compare** numbers with fuzzy matching (IsNumberMatch)
- **Format incrementally** with AsYouTypeFormatter
- **Find numbers in text** with configurable leniency
- **Short numbers** — validate, classify costs, detect emergency numbers
- **Geocoding** — offline geographic descriptions (240+ regions, 34 languages)
- **Timezone mapping** — phone number prefix to IANA timezone
- **Example numbers** — valid numbers by region, with type preference

## Install

```bash
go get github.com/shepard-labs/go-phonesafe
```

Requires Go 1.26+.

## Quick Start

```go
package main

import (
    "fmt"
    phonesafe "github.com/shepard-labs/go-phonesafe"
)

func main() {
    u := phonesafe.Instance()

    // Parse
    num, err := u.Parse("+41 44 668 1800", "CH")
    if err != nil {
        panic(err)
    }

    // Validate
    fmt.Println(u.IsValidNumber(num)) // true

    // Format
    fmt.Println(u.Format(num, phonesafe.FormatE164))          // +41446681800
    fmt.Println(u.Format(num, phonesafe.FormatInternational)) // +41 44 668 18 00
    fmt.Println(u.Format(num, phonesafe.FormatNational))      // 044 668 18 00

    // Number type
    fmt.Println(u.GetNumberType(num)) // TypeFixedLine

    // AsYouType formatting
    f := u.NewAsYouTypeFormatter("US")
    fmt.Println(f.InputDigit('6')) // 6
    fmt.Println(f.InputDigit('5')) // 65
    fmt.Println(f.InputDigit('0')) // 650
    fmt.Println(f.InputDigit('2')) // 650-2

    // Fuzzy match
    match := u.IsNumberMatch("+41-44-668-1800", num)
    fmt.Println(match) // MatchTypeExact

    // Find numbers in text
    m := u.FindNumbers("Call +1 650 253 0000 for info.", "US", phonesafe.LeniencyValid, 100)
    for m.HasNext() {
        found := m.Next()
        fmt.Printf("Found: %q at [%d:%d]\n", found.RawString, found.Start, found.End)
    }
}
```

## Sub-packages

| Package | Description |
|---------|-------------|
| [`phonesafe/shortnumber`](https://pkg.go.dev/github.com/shepard-labs/go-phonesafe/shortnumber) | Short number validation, emergency detection, cost classification |
| [`phonesafe/geocoder`](https://pkg.go.dev/github.com/shepard-labs/go-phonesafe/geocoder) | Offline geographic descriptions for phone numbers |
| [`phonesafe/timezone`](https://pkg.go.dev/github.com/shepard-labs/go-phonesafe/timezone) | Phone number prefix to IANA timezone mapping |

### Geocoder

```go
import (
    phonesafe "github.com/shepard-labs/go-phonesafe"
    "github.com/shepard-labs/go-phonesafe/geocoder"
    "golang.org/x/text/language"
)

g := geocoder.Instance()
u := phonesafe.Instance()
num, _ := u.Parse("+1 650 253 0000", "US")
fmt.Println(g.GetDescriptionForNumber(num, language.English)) // Mountain View, CA
```

### Timezone

```go
import (
    phonesafe "github.com/shepard-labs/go-phonesafe"
    "github.com/shepard-labs/go-phonesafe/timezone"
)

m := timezone.Instance()
u := phonesafe.Instance()
num, _ := u.Parse("+1 650 253 0000", "US")
fmt.Println(m.GetTimeZonesForNumber(num)) // [America/Los_Angeles]
```

### Short Numbers

```go
import (
    phonesafe "github.com/shepard-labs/go-phonesafe"
    "github.com/shepard-labs/go-phonesafe/shortnumber"
)

s := shortnumber.Instance()
u := phonesafe.Instance()
num, _ := u.Parse("911", "US")
fmt.Println(s.IsEmergencyNumber("911", "US"))         // true
fmt.Println(s.IsValidShortNumber(num))                // true
fmt.Println(s.GetExpectedCost(num))                   // CostTollFree
```

### Example Numbers

```go
u := phonesafe.Instance()

// Get a valid example number for any region
num, ok := u.GetExampleNumber("GB")
fmt.Println(num) // e.g. +44 7912 345678

// Get an example number for a region, preferring toll-free
num, ok = u.GetExampleNumberForType("US", phonesafe.TypeTollFree)
fmt.Println(num) // e.g. +1 800 234 5678
```

## Comparison

| Feature                           |        go-phonesafe        | libphonenumber (Java) |    nyaruka/phonenumbers     |
|-----------------------------------|:--------------------------:|:---------------------:|:---------------------------:|
| Parse / Format / Validate         |            Yes             |          Yes          |             Yes             |
| Number type classification        |            Yes             |          Yes          |             Yes             |
| AsYouTypeFormatter                |            Yes             |          Yes          |             No              |
| FindNumbers in text               |            Yes             |          Yes          |             No              |
| Short number validation           |            Yes             |          Yes          |           Partial           |
| Short number cost classification  |            Yes             |          Yes          |             No              |
| Geocoding (240+ regions)          |            Yes             |          Yes          |             Yes             |
| Timezone mapping                  |            Yes             |          Yes          |             Yes             |
| Carrier mapping                   |   No (not current data)    |          Yes          |             Yes             |
| Example numbers                   |            Yes             |          Yes          |             Yes             |
| **Package structure**             |  **Modular (4 packages)**  |        Single         |           Single            |
| **Protobuf dependency**           |           **No**           |          No           |             Yes             |
| **Go version**                    |         **1.26+**          |          N/A          |            1.23+            |
| **Dependencies**                  | **golang.org/x/text only** |         None          | protobuf + testify + x/text |
| **Metadata approach**             |   Go literals (code-gen)   | Loaded from resources |    Binary serialization     |
| `NormalizeDiallableCharsOnly`     |       Yes (exported)       |          Yes          |    No (unexported only)     |
| `IsNumberGeographical`            |       Yes (exported)       |          Yes          |    No (unexported only)     |
| `GetSupportedTypesForRegion`      |            Yes             |          Yes          |             No              |

go-phonesafe is the only Go port with a modular package structure, the fewest dependencies, and full implementations of all core features.

**Why no carrier mapping?** Carrier data goes stale the moment a number is ported (MNP). The data describes the *original carrier* assigned a range, not the current one — making it unreliable for most use cases. If real-time carrier data is needed, integrate a dedicated MNP lookup service.

## Benchmarks

Apple M4 Max, Go 1.26:

| Operation | ns/op | allocs/op |
|-----------|------:|----------:|
| Parse | 5,182 | 7 |
| Format (E.164) | 37 | 2 |
| Format (International) | 516 | 10 |
| Format (National) | 663 | 11 |
| IsNumberMatch | 16 | 0 |
| IsValidNumber | 1,368 | 12 |
| GetNumberType | 1,296 | 12 |
| IsPossibleNumber | 146 | 5 |
| AsYouTypeFormatter | 4,159 | 76 |
| FindNumbers (2 matches) | 59,135 | 90 |

Sub-packages (Apple M4 Max):

| Operation | ns/op | allocs/op |
|-----------|------:|----------:|
| Geocode | 2,708 | 28 |
| Timezone | 3,108 | 29 |
| GetTimeZonesForNumber | 247 | 5 |

```bash
go test -bench=. -benchmem ./...
```

Phone numbers are deceptively complex. See [FALSEHOODS.md](./FALSEHOODS.md) before making assumptions.

## Design Decisions

- **Native Go structs** — no protobuf dependency
- **Code-generated metadata** — upstream XML converted to Go source at build time for zero-cost init
- **RE2 regex** — stdlib `regexp` package, guaranteed linear time
- **Goroutine-safe** — singleton pattern with `sync.Once`; immutable metadata after initialization
- **Minimal dependencies** — only `golang.org/x/text/language` for geocoder locale support

## Updating Metadata

```bash
make generate LIBPHONENUMBER_PATH=../libphonenumber
```

## License

Apache 2.0
