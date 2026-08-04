# Falsehoods Programmers Believe About Phone Numbers

*Adapted from [upstream](https://github.com/google/libphonenumber/blob/master/FALSEHOODS.md) with Go-specific tips for go-phonesafe users.*

---

## The List

1. **An individual has a phone number.** Some don't. Never require a phone number unless essential.

2. **You can make a call to any phone number.** Some devices (EFTPOS, fax, dongles) can't receive calls. Some users can't use voice calls (hearing disabilities, noisy environments).

3. **An individual has only one phone number.** Obviously false.

4. **A phone number uniquely identifies an individual.** Households share numbers. Businesses have multiple inbound lines on one number.

5. **Phone numbers cannot be re-used.** Old numbers get recycled and reassigned.

6. **Phone numbers that are valid today will always be valid.** Numbers are disconnected. Ranges are reassigned. Digits are inserted.
   > **go-phonesafe tip:** Never cache validation results. Call `IsValidNumber()` when you need it fresh.

7. **Each country calling code corresponds to exactly one country.** US/CA/Caribbean share +1. Russia/Kazakhstan share +7.

8. **Each country has only one country calling code.** Kosovo was reachable via +381, +386, or +377 depending on the number.
   > **go-phonesafe tip:** Use `GetRegionCodesForCountryCode()` to see all regions sharing a code.

9. **A phone number is dialable from anywhere.** Some numbers are domestic-only. Some are carrier-specific.
   > **go-phonesafe tip:** Use `CanBeInternationallyDialled()` to check.

10. **You can send a text message to any phone number.** Fixed-line numbers typically can't receive SMS.

11. **Only mobile phones can receive text messages.** Some fixed-line services and VoIP providers support SMS.

12. **There are only two ways to dial a phone number: domestically and from overseas.** Carrier codes (Brazil), mobile/fixed prefixes (Nepal), local vs area dialing (New Zealand) add complexity.
    > **go-phonesafe tip:** Use `FormatNumberForMobileDialing()` to get what a user should actually dial.

13. **To make a number dialable, you only need to change the prefix.** Argentina inserts "15" *after* the area code for domestic mobile dialing.

14. **No prefix of a valid phone number can be a valid phone number.** In some countries, dialing more digits reaches a different endpoint.

15. **An invalid number will not reach an endpoint.** Extra digits are ignored in some countries. Carriers "fix" numbers.

16. **All valid phone numbers follow the ITU specifications.** ITU says max 15 digits. Germany has longer valid numbers.

17. **All valid phone numbers belong to a country.** Non-geographical entities exist (+800 International Freephone, +870 Inmarsat, etc.).

18. **Phone numbers contain only digits.** Israel has `*` prefixes. New Zealand uses `*555`. Vanity numbers use letters.

19. **Phone numbers are always written in ASCII.** Egyptian numbers are commonly written in native digits.
    > **go-phonesafe tip:** `Parse()` handles Arabic-indic and wide-ASCII digit normalization automatically.

20. **A leading zero can always be discarded when dialing from abroad.** Italy's leading zero is part of the number and must be included internationally.
    > **go-phonesafe tip:** The library handles this via `ItalianLeadingZero`. Never strip leading zeros yourself.

21. **The country or area code indicates the user's location, language, or timezone.** People move, keep numbers, use VoIP from other countries.
    > **go-phonesafe tip:** Geocoding gives the *number's origin area*, not the person's current location.

22. **The plus sign is optional or can always be replaced by `00`.** The IDD prefix varies by country. `00` is common but not universal.
    > **go-phonesafe tip:** Always store numbers in E.164 format (`Format(number, FormatE164)`). This includes the `+`.

23. **Users will only store phone numbers in phone number fields.** People store birthdays and notes in contact fields.

24. **Phone numbers are numbers.** Never store as `int`. 007 ≠ 7 for phone numbers. Use `string` for storage (or `PhoneNumber` struct for processing).
    > **go-phonesafe tip:** Store as E.164 string in your database. Parse on retrieval when you need to operate on it.

25. **Phone numbering plans represent reality.** Plans are published before, during, or after real-world activation.

---

## Go-Specific Anti-Patterns

### Don't do this:

```go
// BAD: Storing as integer
type User struct {
    Phone int64  // WRONG: loses leading zeros, limits length
}

// BAD: Validating with regex
var phoneRegex = regexp.MustCompile(`^\+?[0-9]{10,15}$`)  // WRONG: too simplistic

// BAD: Assuming format
func formatPhone(n string) string {
    return fmt.Sprintf("(%s) %s-%s", n[:3], n[3:6], n[6:])  // WRONG: US-only assumption
}

// BAD: Caching validation
type CachedNumber struct {
    Number  string
    IsValid bool  // WRONG: validity changes over time with metadata updates
}
```

### Do this instead:

```go
import "github.com/shepard-labs/go-phonesafe"

// GOOD: Parse and validate properly
u := phonesafe.Instance()
num, err := u.Parse(rawInput, "US")
if err != nil {
    // Handle: ErrNotANumber, ErrTooShortNSN, etc.
}

if !u.IsValidNumber(num) {
    // Reject invalid number
}

// GOOD: Store as E.164 string
dbPhone := u.Format(num, phonesafe.FormatE164)  // "+14155551234"

// GOOD: Format for display based on context
display := u.Format(num, phonesafe.FormatNational)        // "(415) 555-1234"
intl := u.Format(num, phonesafe.FormatInternational)      // "+1 415 555 1234"
dial := u.FormatNumberForMobileDialing(num, "GB", true)   // "001 1 415 555 1234"
```
