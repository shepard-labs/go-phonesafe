package geocoder_test

import (
	"fmt"

	phonesafe "github.com/shepard-labs/go-phonesafe"
	"github.com/shepard-labs/go-phonesafe/geocoder"
	"golang.org/x/text/language"
)

func ExampleGeocoder_GetDescriptionForNumber() {
	g := geocoder.Instance()
	u := phonesafe.Instance()
	num, _ := u.Parse("+1 650 253 0000", "US")
	desc := g.GetDescriptionForNumber(num, language.English)
	fmt.Println(desc)
	// Output: Mountain View, CA
}
