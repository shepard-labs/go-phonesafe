package phonesafe_test

import (
	"fmt"

	phonesafe "github.com/shepard-labs/go-phonesafe"
)

func ExamplePhoneNumberUtil_Parse() {
	u := phonesafe.Instance()
	num, err := u.Parse("+41 44 668 1800", "CH")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("CC=%d National=%d\n", num.CountryCode, num.NationalNumber)
	// Output: CC=41 National=446681800
}

func ExamplePhoneNumberUtil_Format() {
	u := phonesafe.Instance()
	num, _ := u.Parse("+1 650 253 0000", "US")
	fmt.Println(u.Format(num, phonesafe.FormatE164))
	fmt.Println(u.Format(num, phonesafe.FormatInternational))
	fmt.Println(u.Format(num, phonesafe.FormatNational))
	fmt.Println(u.Format(num, phonesafe.FormatRFC3966))
	// Output:
	// +16502530000
	// +1 650-253-0000
	// (650) 253-0000
	// tel:+1-650-253-0000
}

func ExamplePhoneNumberUtil_IsValidNumber() {
	u := phonesafe.Instance()
	num, _ := u.Parse("+1 650 253 0000", "US")
	fmt.Println(u.IsValidNumber(num))

	invalid := phonesafe.PhoneNumber{CountryCode: 1, NationalNumber: 123}
	fmt.Println(u.IsValidNumber(invalid))
	// Output:
	// true
	// false
}

func ExamplePhoneNumberUtil_FindNumbers() {
	u := phonesafe.Instance()
	m := u.FindNumbers("Call +1 650 253 0000 for info.", "US", phonesafe.LeniencyValid, 100)
	for m.HasNext() {
		match := m.Next()
		fmt.Printf("Found: %q at [%d:%d]\n", match.RawString, match.Start, match.End)
	}
	// Output: Found: "+1 650 253 0000" at [5:20]
}

func ExampleAsYouTypeFormatter() {
	u := phonesafe.Instance()
	f := u.NewAsYouTypeFormatter("US")
	fmt.Println(f.InputDigit('6'))
	fmt.Println(f.InputDigit('5'))
	fmt.Println(f.InputDigit('0'))
	fmt.Println(f.InputDigit('2'))
	fmt.Println(f.InputDigit('5'))
	fmt.Println(f.InputDigit('3'))
	// Output:
	// 6
	// 65
	// 650
	// 650-2
	// 650-25
	// 650-253
}
