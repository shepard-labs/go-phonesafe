package timezone_test

import (
	"fmt"

	phonesafe "github.com/shepard-labs/go-phonesafe"
	"github.com/shepard-labs/go-phonesafe/timezone"
)

func ExampleMapper_GetTimeZonesForNumber() {
	m := timezone.Instance()
	u := phonesafe.Instance()
	num, _ := u.Parse("+1 650 253 0000", "US")
	zones := m.GetTimeZonesForNumber(num)
	fmt.Println(zones)
	// Output: [America/Los_Angeles]
}
