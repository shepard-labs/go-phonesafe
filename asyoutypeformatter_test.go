package phonesafe

import "testing"

// step represents a single input→expected output pair for AYTF testing.
type step struct {
	input rune
	want  string
}

func runSequence(t *testing.T, f *AsYouTypeFormatter, steps []step) {
	t.Helper()
	for i, s := range steps {
		got := f.InputDigit(s.input)
		if got != s.want {
			t.Errorf("step %d: InputDigit(%q) = %q, want %q", i, string(s.input), got, s.want)
		}
	}
}

func TestAYTF_US(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("US")

	// US domestic: (650) 253-2222
	runSequence(t, f, []step{
		{'6', "6"},
		{'5', "65"},
		{'0', "650"},
		{'2', "650-2"},
		{'5', "650-25"},
		{'3', "650-253"},
		{'2', "650-2532"},
		{'2', "(650) 253-22"},
		{'2', "(650) 253-222"},
		{'2', "(650) 253-2222"},
	})

	// US with national prefix 1: 1 (650) 253-2222
	f.Clear()
	runSequence(t, f, []step{
		{'1', "1"},
		{'6', "16"},
		{'5', "1 65"},
		{'0', "1 (650"},
		{'2', "1 (650) 2"},
		{'5', "1 (650) 25"},
		{'3', "1 (650) 253"},
		{'2', "1 (650) 253-2"},
		{'2', "1 (650) 253-22"},
		{'2', "1 (650) 253-222"},
		{'2', "1 (650) 253-2222"},
	})

	// US IDD to UK London: 011 44 20 7031 3000
	f.Clear()
	runSequence(t, f, []step{
		{'0', "0"},
		{'1', "01"},
		{'1', "011 "},
		{'4', "011 4"},
		{'4', "011 44 "},
		{'2', "011 44 2"},
		{'0', "011 44 20"},
		{'7', "011 44 20 7"},
		{'0', "011 44 20 70"},
		{'3', "011 44 20 703"},
		{'1', "011 44 20 7031"},
		{'3', "011 44 20 7031 3"},
		{'0', "011 44 20 7031 30"},
		{'0', "011 44 20 7031 300"},
		{'0', "011 44 20 7031 3000"},
	})

	// US IDD to Argentina mobile: 011 54 9 11 2312-1234
	f.Clear()
	runSequence(t, f, []step{
		{'0', "0"},
		{'1', "01"},
		{'1', "011 "},
		{'5', "011 5"},
		{'4', "011 54 "},
		{'9', "011 54 9"},
		{'1', "011 54 91"},
		{'1', "011 54 9 11"},
		{'2', "011 54 9 11 2"},
		{'3', "011 54 9 11 23"},
		{'1', "011 54 9 11 231"},
		{'2', "011 54 9 11 2312"},
		{'1', "011 54 9 11 2312-1"},
		{'2', "011 54 9 11 2312-12"},
		{'3', "011 54 9 11 2312-123"},
		{'4', "011 54 9 11 2312-1234"},
	})

	// US IDD to Angola: 011 244 280 000 000
	f.Clear()
	runSequence(t, f, []step{
		{'0', "0"},
		{'1', "01"},
		{'1', "011 "},
		{'2', "011 2"},
		{'4', "011 24"},
		{'4', "011 244 "},
		{'2', "011 244 2"},
		{'8', "011 244 28"},
		{'0', "011 244 280"},
		{'0', "011 244 280 0"},
		{'0', "011 244 280 00"},
		{'0', "011 244 280 000"},
		{'0', "011 244 280 000 0"},
		{'0', "011 244 280 000 00"},
		{'0', "011 244 280 000 000"},
	})

	// US with +PL number
	f.Clear()
	runSequence(t, f, []step{
		{'+', "+"},
		{'4', "+4"},
		{'8', "+48 "},
		{'8', "+48 8"},
		{'8', "+48 88"},
		{'1', "+48 881"},
		{'2', "+48 881 2"},
		{'3', "+48 881 23"},
		{'1', "+48 881 231"},
		{'2', "+48 881 231 2"},
		{'1', "+48 881 231 21"},
		{'2', "+48 881 231 212"},
	})
}

func TestAYTF_USFullWidthCharacters(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("US")

	// Full-width digits should be normalized; first 2 are raw, then formatting kicks in.
	runSequence(t, f, []step{
		{'６', "６"},
		{'５', "６５"},
		{'０', "650"},
		{'２', "650-2"},
		{'５', "650-25"},
		{'３', "650-253"},
		{'２', "650-2532"},
		{'２', "(650) 253-22"},
		{'２', "(650) 253-222"},
		{'２', "(650) 253-2222"},
	})
}

func TestAYTF_USMobileShortCode(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("US")

	// Short codes with * and # — user formatting, no auto-format.
	runSequence(t, f, []step{
		{'*', "*"},
		{'1', "*1"},
		{'2', "*12"},
		{'1', "*121"},
		{'#', "*121#"},
	})
}

func TestAYTF_USVanityNumber(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("US")

	// User-formatted vanity: once space is entered, formatting stops.
	runSequence(t, f, []step{
		{'8', "8"},
		{'0', "80"},
		{'0', "800"},
		{' ', "800 "},
		{'M', "800 M"},
		{'Y', "800 MY"},
		{' ', "800 MY "},
		{'A', "800 MY A"},
		{'P', "800 MY AP"},
		{'P', "800 MY APP"},
		{'L', "800 MY APPL"},
		{'E', "800 MY APPLE"},
	})
}

func TestAYTF_GBFixedLine(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("GB")

	// London number: 020 7031 3000
	runSequence(t, f, []step{
		{'0', "0"},
		{'2', "02"},
		{'0', "020"},
		{'7', "020 7"},
		{'0', "020 70"},
		{'3', "020 703"},
		{'1', "020 7031"},
		{'3', "020 7031 3"},
		{'0', "020 7031 30"},
		{'0', "020 7031 300"},
		{'0', "020 7031 3000"},
	})
}

func TestAYTF_GBTollFree(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("GB")

	// 0800 number (production metadata: 0800 NNN NNNN)
	runSequence(t, f, []step{
		{'0', "0"},
		{'8', "08"},
		{'0', "080"},
		{'0', "0800"},
		{'7', "0800 7"},
		{'0', "0800 70"},
		{'3', "0800 703"},
		{'1', "0800 7031"},
		{'3', "0800 70313"},
		{'0', "0800 703130"},
		{'0', "0800 703 1300"},
	})
}

func TestAYTF_DE(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("DE")

	// Berlin number: 030 1234...
	runSequence(t, f, []step{
		{'0', "0"},
		{'3', "03"},
		{'0', "030"},
		{'1', "030 1"},
		{'2', "030 12"},
		{'3', "030 123"},
		{'4', "030 1234"},
		{'5', "030 12345"},
		{'6', "030 123456"},
		{'7', "030 1234567"},
		{'8', "030 12345678"},
	})

	// DE IDD to US: 00 1 650-253-2222
	f.Clear()
	runSequence(t, f, []step{
		{'0', "0"},
		{'0', "00"},
		{'1', "00 1 "},
		{'6', "00 1 6"},
		{'5', "00 1 65"},
		{'0', "00 1 650"},
		{'2', "00 1 650-2"},
		{'5', "00 1 650-25"},
		{'3', "00 1 650-253"},
		{'2', "00 1 650-253-2"},
		{'2', "00 1 650-253-22"},
		{'2', "00 1 650-253-222"},
		{'2', "00 1 650-253-2222"},
	})
}

func TestAYTF_AR(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("AR")

	// Argentina domestic Buenos Aires: 011 7031-3000
	runSequence(t, f, []step{
		{'0', "0"},
		{'1', "01"},
		{'1', "011"},
		{'7', "011 7"},
		{'0', "011 70"},
		{'3', "011 703"},
		{'1', "011 7031"},
		{'3', "011 7031-3"},
		{'0', "011 7031-30"},
		{'0', "011 7031-300"},
		{'0', "011 7031-3000"},
	})

	// Argentina mobile international: +54 9 11 2312-1234
	f.Clear()
	runSequence(t, f, []step{
		{'+', "+"},
		{'5', "+5"},
		{'4', "+54 "},
		{'9', "+54 9"},
		{'1', "+54 91"},
		{'1', "+54 9 11"},
		{'2', "+54 9 11 2"},
		{'3', "+54 9 11 23"},
		{'1', "+54 9 11 231"},
		{'2', "+54 9 11 2312"},
		{'1', "+54 9 11 2312-1"},
		{'2', "+54 9 11 2312-12"},
		{'3', "+54 9 11 2312-123"},
		{'4', "+54 9 11 2312-1234"},
	})
}

func TestAYTF_KR(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("KR")

	// +82 51-234-5678 (Busan)
	runSequence(t, f, []step{
		{'+', "+"},
		{'8', "+8"},
		{'2', "+82 "},
		{'5', "+82 5"},
		{'1', "+82 51"},
		{'2', "+82 51-2"},
		{'3', "+82 51-23"},
		{'4', "+82 51-234"},
		{'5', "+82 51-2345"},
		{'6', "+82 51-2345-6"},
		{'7', "+82 51-2345-67"},
		{'8', "+82 51-234-5678"},
	})

	// +82 2-531-5678 (Seoul)
	f.Clear()
	runSequence(t, f, []step{
		{'+', "+"},
		{'8', "+8"},
		{'2', "+82 "},
		{'2', "+82 2"},
		{'5', "+82 25"},
		{'3', "+82 2-53"},
		{'1', "+82 2-531"},
		{'5', "+82 2-5315"},
		{'6', "+82 2-5315-6"},
		{'7', "+82 2-5315-67"},
		{'8', "+82 2-531-5678"},
	})

	// KR domestic: 011-456-7890
	f.Clear()
	runSequence(t, f, []step{
		{'0', "0"},
		{'1', "01"},
		{'1', "011"},
		{'4', "011-4"},
		{'5', "011-45"},
		{'6', "011-456"},
		{'7', "011-4567"},
		{'8', "011-4567-8"},
		{'9', "011-4567-89"},
		{'0', "011-456-7890"},
	})
}

func TestAYTF_MX(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("MX")

	// +52 800 123 4567 (toll-free)
	runSequence(t, f, []step{
		{'+', "+"},
		{'5', "+5"},
		{'2', "+52 "},
		{'8', "+52 8"},
		{'0', "+52 80"},
		{'0', "+52 800"},
		{'1', "+52 800 1"},
		{'2', "+52 800 12"},
		{'3', "+52 800 123"},
		{'4', "+52 800 123 4"},
		{'5', "+52 800 123 45"},
		{'6', "+52 800 123 456"},
		{'7', "+52 800 123 4567"},
	})

	// +52 55 1234 5678 (Mexico City)
	f.Clear()
	runSequence(t, f, []step{
		{'+', "+"},
		{'5', "+5"},
		{'2', "+52 "},
		{'5', "+52 5"},
		{'5', "+52 55"},
		{'1', "+52 55 1"},
		{'2', "+52 55 12"},
		{'3', "+52 55 123"},
		{'4', "+52 55 1234"},
		{'5', "+52 55 1234 5"},
		{'6', "+52 55 1234 56"},
		{'7', "+52 55 1234 567"},
		{'8', "+52 55 1234 5678"},
	})
}

func TestAYTF_JP(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("JP")

	// +81 50-2345-6789 (IP phone)
	runSequence(t, f, []step{
		{'+', "+"},
		{'8', "+8"},
		{'1', "+81 "},
		{'5', "+81 5"},
		{'0', "+81 50"},
		{'2', "+81 50-2"},
		{'3', "+81 50-23"},
		{'4', "+81 50-234"},
		{'5', "+81 50-2345"},
		{'6', "+81 50-2345-6"},
		{'7', "+81 50-2345-67"},
		{'8', "+81 50-2345-678"},
		{'9', "+81 50-2345-6789"},
	})
}

func TestAYTF_LongIDD_AU(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("AU")

	// 0011 1 650-253-2222 (AU IDD to US)
	runSequence(t, f, []step{
		{'0', "0"},
		{'0', "00"},
		{'1', "001"},
		{'1', "0011"},
		{'1', "0011 1 "},
		{'6', "0011 1 6"},
		{'5', "0011 1 65"},
		{'0', "0011 1 650"},
		{'2', "0011 1 650-2"},
		{'5', "0011 1 650-25"},
		{'3', "0011 1 650-253"},
		{'2', "0011 1 650-253-2"},
		{'2', "0011 1 650-253-22"},
		{'2', "0011 1 650-253-222"},
		{'2', "0011 1 650-253-2222"},
	})
}

func TestAYTF_AUDomestic(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("AU")

	// 02 1234 5678 (domestic with national prefix 0)
	runSequence(t, f, []step{
		{'0', "0"},
		{'2', "02"},
		{'1', "021"},
		{'2', "02 12"},
		{'3', "02 123"},
		{'4', "02 1234"},
		{'5', "02 1234 5"},
		{'6', "02 1234 56"},
		{'7', "02 1234 567"},
		{'8', "02 1234 5678"},
	})
}

func TestAYTF_LongNDD_KR(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("KR")

	// 08811-9876-7890 (long NDD prefix)
	// KR national prefix "0" is extracted, then 8811 acts as carrier prefix.
	// Format switches at 10 and 11 digits as longer patterns match.
	runSequence(t, f, []step{
		{'0', "0"},
		{'8', "08"},
		{'8', "088"},
		{'1', "0881"},
		{'1', "08811"},
		{'9', "08811-9"},
		{'8', "08811-98"},
		{'7', "08811-987"},
		{'6', "08811-9876"},
		{'7', "08811-9876-7"},
		{'8', "08811-9876-78"},
		{'9', "08811-987-6789"},  // format switch to 3+4 pattern
		{'0', "08811-9876-7890"}, // switches to 4+4 pattern
	})
}

func TestAYTF_InvalidRegion(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("ZZ")

	// With +, formatting should still work via country code detection.
	runSequence(t, f, []step{
		{'+', "+"},
		{'4', "+4"},
		{'8', "+48 "},
		{'8', "+48 8"},
		{'8', "+48 88"},
		{'1', "+48 881"},
		{'2', "+48 881 2"},
		{'3', "+48 881 23"},
		{'1', "+48 881 231"},
		{'2', "+48 881 231 2"},
	})

	// Without +, no formatting (no default region).
	f.Clear()
	runSequence(t, f, []step{
		{'6', "6"},
		{'5', "65"},
		{'0', "650"},
		{'2', "6502"},
		{'5', "65025"},
		{'3', "650253"},
	})
}

func TestAYTF_InvalidPlusSign(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("ZZ")

	runSequence(t, f, []step{
		{'+', "+"},
		{'4', "+4"},
		{'8', "+48 "},
		{'8', "+48 8"},
		{'8', "+48 88"},
		{'1', "+48 881"},
		{'2', "+48 881 2"},
		{'3', "+48 881 23"},
		{'1', "+48 881 231"},
	})
	// Plus sign in the middle stops formatting.
	got := f.InputDigit('+')
	if got != "+48881231+" {
		t.Errorf("InputDigit('+') = %q, want %q", got, "+48881231+")
	}
	got = f.InputDigit('2')
	if got != "+48881231+2" {
		t.Errorf("InputDigit('2') = %q, want %q", got, "+48881231+2")
	}
}

func TestAYTF_RememberedPosition_US(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("US")

	got := f.InputDigitAndRememberPosition('1')
	if got != "1" {
		t.Errorf("got %q, want %q", got, "1")
	}
	if pos := f.GetRememberedPosition(); pos != 1 {
		t.Errorf("GetRememberedPosition() = %d, want 1", pos)
	}

	f.InputDigit('6')
	f.InputDigit('5')
	if pos := f.GetRememberedPosition(); pos != 1 {
		t.Errorf("GetRememberedPosition() = %d, want 1", pos)
	}

	got = f.InputDigitAndRememberPosition('0')
	if got != "1 (650" {
		t.Errorf("got %q, want %q", got, "1 (650")
	}
	if pos := f.GetRememberedPosition(); pos != 6 {
		t.Errorf("GetRememberedPosition() = %d, want 6", pos)
	}

	f.InputDigit('2')
	f.InputDigit('5')
	if pos := f.GetRememberedPosition(); pos != 6 {
		t.Errorf("GetRememberedPosition() = %d, want 6", pos)
	}

	f.InputDigit('3')
	f.InputDigit('2')
	f.InputDigit('2')

	got = f.InputDigitAndRememberPosition('2')
	if got != "1 (650) 253-222" {
		t.Errorf("got %q, want %q", got, "1 (650) 253-222")
	}
	if pos := f.GetRememberedPosition(); pos != 15 {
		t.Errorf("GetRememberedPosition() = %d, want 15", pos)
	}

	got = f.InputDigit('2')
	if got != "1 (650) 253-2222" {
		t.Errorf("got %q, want %q", got, "1 (650) 253-2222")
	}
	if pos := f.GetRememberedPosition(); pos != 15 {
		t.Errorf("GetRememberedPosition() = %d, want 15", pos)
	}

	// Too many digits — formatting drops.
	got = f.InputDigit('2')
	if got != "165025322222" {
		t.Errorf("got %q, want %q", got, "165025322222")
	}
	if pos := f.GetRememberedPosition(); pos != 10 {
		t.Errorf("GetRememberedPosition() = %d, want 10", pos)
	}
}

func TestAYTF_RememberedPosition_UserFormatting(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("US")

	f.InputDigit('6')
	f.InputDigit('5')
	f.InputDigit('0')
	// User enters dash — formatting stops.
	got := f.InputDigit('-')
	if got != "650-" {
		t.Errorf("got %q, want %q", got, "650-")
	}

	got = f.InputDigitAndRememberPosition('2')
	if got != "650-2" {
		t.Errorf("got %q, want %q", got, "650-2")
	}
	if pos := f.GetRememberedPosition(); pos != 5 {
		t.Errorf("GetRememberedPosition() = %d, want 5", pos)
	}

	f.InputDigit('5')
	if pos := f.GetRememberedPosition(); pos != 5 {
		t.Errorf("GetRememberedPosition() = %d, want 5", pos)
	}
}

func TestAYTF_International_Toll_Free(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("US")

	// +800 1234 5678
	runSequence(t, f, []step{
		{'+', "+"},
		{'8', "+8"},
		{'0', "+80"},
		{'0', "+800 "},
		{'1', "+800 1"},
		{'2', "+800 12"},
		{'3', "+800 123"},
		{'4', "+800 1234"},
		{'5', "+800 1234 5"},
		{'6', "+800 1234 56"},
		{'7', "+800 1234 567"},
		{'8', "+800 1234 5678"},
	})
}

func TestAYTF_TooLongNumber(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("ZZ")

	// +81 numbers that exceed format length should degrade to unformatted.
	runSequence(t, f, []step{
		{'+', "+"},
		{'8', "+8"},
		{'1', "+81 "},
		{'9', "+81 9"},
		{'0', "+81 90"},
		{'1', "+81 90-1"},
		{'2', "+81 90-12"},
		{'3', "+81 90-123"},
		{'4', "+81 90-1234"},
		{'5', "+81 90-1234-5"},
		{'6', "+81 90-1234-56"},
		{'7', "+81 90-1234-567"},
		{'8', "+81 90-1234-5678"},
		{'9', "+8190123456789"},
		{'0', "+81901234567890"},
		{'1', "+819012345678901"},
	})
}

func TestAYTF_BY(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("BY")

	runSequence(t, f, []step{
		{'8', "8"},
		{'8', "88"},
		{'1', "881"},
		{'9', "8 819"},
		{'0', "8 819 0"},
		{'1', "8 819 01"},
		{'2', "8 819 012"},
		{'3', "8 819 0123"},
	})
}

func TestAYTF_Clear(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("US")

	// Format a number, then clear and format another.
	f.InputDigit('6')
	f.InputDigit('5')
	f.InputDigit('0')
	f.InputDigit('2')
	f.InputDigit('5')
	f.InputDigit('3')
	f.InputDigit('2')
	f.InputDigit('2')
	f.InputDigit('2')
	got := f.InputDigit('2')
	if got != "(650) 253-2222" {
		t.Errorf("before clear: got %q, want %q", got, "(650) 253-2222")
	}

	f.Clear()
	got = f.InputDigit('1')
	if got != "1" {
		t.Errorf("after clear: got %q, want %q", got, "1")
	}
	got = f.InputDigit('6')
	if got != "16" {
		t.Errorf("after clear: got %q, want %q", got, "16")
	}
	got = f.InputDigit('5')
	if got != "1 65" {
		t.Errorf("after clear: got %q, want %q", got, "1 65")
	}
}

func TestAYTF_RememberedPosition_International(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("US")

	// +1 6... with remembered position
	got := f.InputDigit('+')
	if got != "+" {
		t.Errorf("got %q, want %q", got, "+")
	}
	f.InputDigit('1')
	got = f.InputDigitAndRememberPosition('6')
	if got != "+1 6" {
		t.Errorf("got %q, want %q", got, "+1 6")
	}
	f.InputDigit('5')
	f.InputDigit('0')
	if pos := f.GetRememberedPosition(); pos != 4 {
		t.Errorf("GetRememberedPosition() = %d, want 4", pos)
	}
	f.InputDigit('2')
	if pos := f.GetRememberedPosition(); pos != 4 {
		t.Errorf("GetRememberedPosition() = %d, want 4", pos)
	}
}

func TestAYTF_GBRememberedPosition(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("GB")

	f.InputDigit('0')
	f.InputDigit('2')
	f.InputDigit('0')
	got := f.InputDigitAndRememberPosition('7')
	if got != "020 7" {
		t.Errorf("got %q, want %q", got, "020 7")
	}
	if pos := f.GetRememberedPosition(); pos != 5 {
		t.Errorf("GetRememberedPosition() = %d, want 5", pos)
	}
	f.InputDigit('0')
	f.InputDigit('3')
	if pos := f.GetRememberedPosition(); pos != 5 {
		t.Errorf("GetRememberedPosition() = %d, want 5", pos)
	}
	f.InputDigit('1')
	got = f.InputDigit('3')
	if got != "020 7031 3" {
		t.Errorf("got %q, want %q", got, "020 7031 3")
	}
}

// Upstream: AsYouTypeFormatterTest.testAYTFGBPremiumRate
func TestAYTF_GBPremiumRate(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("GB")

	// 090 7031 3000
	runSequence(t, f, []step{
		{'0', "0"},
		{'9', "09"},
		{'0', "090"},
		{'7', "0907"},
		{'0', "0907 0"},
		{'3', "0907 03"},
		{'1', "0907 031"},
		{'3', "0907 031 3"},
		{'0', "0907 031 30"},
		{'0', "0907 031 300"},
		{'0', "0907 031 3000"},
	})
}

// Upstream: AsYouTypeFormatterTest.testAYTFNZMobile
func TestAYTF_NZMobile(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("NZ")

	// 021 123 456 — note: Go uses space separator, upstream uses hyphen.
	runSequence(t, f, []step{
		{'0', "0"},
		{'2', "02"},
		{'1', "021"},
		{'1', "021 1"},
		{'2', "021 12"},
		{'3', "021 123"},
		{'4', "021 123 4"},
		{'5', "021 123 45"},
		{'6', "021 123 456"},
	})
}

// Upstream: AsYouTypeFormatterTest.testAYTFARMobile
func TestAYTF_ARMobile(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("AR")

	// +54 9 11 2312-1234 — note: Go keeps hyphen in local number, upstream uses space.
	runSequence(t, f, []step{
		{'+', "+"},
		{'5', "+5"},
		{'4', "+54 "},
		{'9', "+54 9"},
		{'1', "+54 91"},
		{'1', "+54 9 11"},
		{'2', "+54 9 11 2"},
		{'3', "+54 9 11 23"},
		{'1', "+54 9 11 231"},
		{'2', "+54 9 11 2312"},
		{'1', "+54 9 11 2312-1"},
		{'2', "+54 9 11 2312-12"},
		{'3', "+54 9 11 2312-123"},
		{'4', "+54 9 11 2312-1234"},
	})
}

// Upstream: AsYouTypeFormatterTest.testAYTFMultipleLeadingDigitPatterns
func TestAYTF_MultipleLeadingDigitPatterns(t *testing.T) {
	u := Instance()

	// +81 50 2345 6789 — note: Go uses hyphen in local number, upstream uses space.
	f := u.NewAsYouTypeFormatter("JP")
	runSequence(t, f, []step{
		{'+', "+"},
		{'8', "+8"},
		{'1', "+81 "},
		{'5', "+81 5"},
		{'0', "+81 50"},
		{'2', "+81 50-2"},
		{'3', "+81 50-23"},
		{'4', "+81 50-234"},
		{'5', "+81 50-2345"},
		{'6', "+81 50-2345-6"},
		{'7', "+81 50-2345-67"},
		{'8', "+81 50-2345-678"},
		{'9', "+81 50-2345-6789"},
	})

	// +81 222 12 5678 — note: Go uses hyphen in local number, upstream uses space.
	f = u.NewAsYouTypeFormatter("JP")
	runSequence(t, f, []step{
		{'+', "+"},
		{'8', "+8"},
		{'1', "+81 "},
		{'2', "+81 2"},
		{'2', "+81 22"},
		{'2', "+81 22-2"},
		{'1', "+81 22-21"},
		{'2', "+81 22-212"},
		{'5', "+81 22-212-5"},
		{'6', "+81 22-212-56"},
		{'7', "+81 22-212-567"},
		{'8', "+81 22-212-5678"},
	})

	// 011113 — note: Go uses hyphen for local number, upstream uses space.
	f = u.NewAsYouTypeFormatter("JP")
	runSequence(t, f, []step{
		{'0', "0"},
		{'1', "01"},
		{'1', "011"},
		{'1', "011-1"},
		{'1', "011-11"},
		{'3', "011-113"},
	})

	// +81 3332 2 5678 — note: Go uses hyphen in local number, upstream uses space.
	f = u.NewAsYouTypeFormatter("JP")
	runSequence(t, f, []step{
		{'+', "+"},
		{'8', "+8"},
		{'1', "+81 "},
		{'3', "+81 3"},
		{'3', "+81 33"},
		{'3', "+81 3-33"},
		{'2', "+81 3-332"},
		{'2', "+81 3-3322"},
		{'5', "+81 3-3322-5"},
		{'6', "+81 3-3322-56"},
		{'7', "+81 3-3322-567"},
		{'8', "+81 3-3322-5678"},
	})
}

// Upstream: AsYouTypeFormatterTest.testAYTFLongIDD_KR
func TestAYTF_LongIDD_KR(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("KR")

	// 00300 1 650 253 2222 — Go uses compact formatting after IDD.
	runSequence(t, f, []step{
		{'0', "0"},
		{'0', "00"},
		{'3', "003"},
		{'0', "0030"},
		{'0', "00300"},
		{'1', "003001"},
		{'6', "0030016"},
		{'5', "00300165"},
		{'0', "003001650"},
		{'2', "0030016502"},
		{'5', "00300165025"},
		{'3', "003001650253"},
		{'2', "0030016502532"},
		{'2', "00300165025322"},
		{'2', "003001650253222"},
		{'2', "0030016502532222"},
	})
}

// Upstream: AsYouTypeFormatterTest.testAYTFLongNDD_SG
func TestAYTF_LongNDD_SG(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("SG")

	// 777777 9876 7890 — note: Go uses different grouping pattern.
	runSequence(t, f, []step{
		{'7', "7"},
		{'7', "77"},
		{'7', "777"},
		{'7', "7777"},
		{'7', "77777"},
		{'7', "7777 77"},
		{'9', "7777 779"},
		{'8', "7777 7798"},
		{'7', "7777 7798 7"},
		{'6', "7777 7798 76"},
		{'7', "7777 7798 767"},
		{'8', "777777987678"},
		{'9', "7777779876789"},
		{'0', "77777798767890"},
	})
}

// Upstream: AsYouTypeFormatterTest.testAYTFShortNumberFormattingFix_AU
func TestAYTF_ShortNumberFormattingFix_AU(t *testing.T) {
	u := Instance()

	// 1234567890 — Go uses no prefix group for AU short numbers without +.
	f := u.NewAsYouTypeFormatter("AU")
	runSequence(t, f, []step{
		{'1', "1"},
		{'2', "12"},
		{'3', "123"},
		{'4', "1234"},
		{'5', "12345"},
		{'6', "123456"},
		{'7', "1234567"},
		{'8', "12345678"},
		{'9', "123456789"},
		{'0', "1234567890"},
	})

	// +61 1234 567 890 — note: Go adds space after +61 before national number.
	f = u.NewAsYouTypeFormatter("AU")
	runSequence(t, f, []step{
		{'+', "+"},
		{'6', "+6"},
		{'1', "+61 "},
		{'1', "+61 1"},
		{'2', "+61 12"},
		{'3', "+61123"},
		{'4', "+611234"},
		{'5', "+6112345"},
		{'6', "+61123456"},
		{'7', "+611234567"},
		{'8', "+6112345678"},
		{'9', "+61123456789"},
		{'0', "+611234567890"},
	})

	// 0212345678 - For leading digit 2, the national prefix formatting rule puts the national prefix before the first group.
	f = u.NewAsYouTypeFormatter("AU")
	runSequence(t, f, []step{
		{'0', "0"},
		{'2', "02"},
		{'1', "021"},
		{'2', "02 12"},
		{'3', "02 123"},
		{'4', "02 1234"},
		{'5', "02 1234 5"},
		{'6', "02 1234 56"},
		{'7', "02 1234 567"},
		{'8', "02 1234 5678"},
	})

	// 212345678 - Without the leading 0.
	f = u.NewAsYouTypeFormatter("AU")
	runSequence(t, f, []step{
		{'2', "2"},
		{'1', "21"},
		{'2', "212"},
		{'3', "2123"},
		{'4', "21234"},
		{'5', "212345"},
		{'6', "2123456"},
		{'7', "21234567"},
		{'8', "212345678"},
	})

	// +61 2 1234 5678
	f = u.NewAsYouTypeFormatter("AU")
	runSequence(t, f, []step{
		{'+', "+"},
		{'6', "+6"},
		{'1', "+61 "},
		{'2', "+61 2"},
		{'1', "+61 21"},
		{'2', "+61 2 12"},
		{'3', "+61 2 123"},
		{'4', "+61 2 1234"},
		{'5', "+61 2 1234 5"},
		{'6', "+61 2 1234 56"},
		{'7', "+61 2 1234 567"},
		{'8', "+61 2 1234 5678"},
	})
}

// Upstream: AsYouTypeFormatterTest.testAYTFShortNumberFormattingFix_KR
func TestAYTF_ShortNumberFormattingFix_KR(t *testing.T) {
	u := Instance()

	// 111
	f := u.NewAsYouTypeFormatter("KR")
	runSequence(t, f, []step{
		{'1', "1"},
		{'1', "11"},
		{'1', "111"},
	})

	// 114
	f = u.NewAsYouTypeFormatter("KR")
	runSequence(t, f, []step{
		{'1', "1"},
		{'1', "11"},
		{'4', "114"},
	})

	// 13121234 - Mobile number without national prefix; formatted with hyphen.
	f = u.NewAsYouTypeFormatter("KR")
	runSequence(t, f, []step{
		{'1', "1"},
		{'3', "13"},
		{'1', "131"},
		{'2', "1312"},
		{'1', "1312-1"},
		{'2', "1312-12"},
		{'3', "1312-123"},
		{'4', "1312-1234"},
	})

	// +82 131-2-1234
	f = u.NewAsYouTypeFormatter("KR")
	runSequence(t, f, []step{
		{'+', "+"},
		{'8', "+8"},
		{'2', "+82 "},
		{'1', "+82 1"},
		{'3', "+82 13"},
		{'1', "+82 131"},
		{'2', "+82 1312"},
		{'1', "+82 1312-1"},
		{'2', "+82 1312-12"},
		{'3', "+82 1312-123"},
		{'4', "+82 1312-1234"},
	})
}

// Upstream: AsYouTypeFormatterTest.testAYTFShortNumberFormattingFix_MX
func TestAYTF_ShortNumberFormattingFix_MX(t *testing.T) {
	u := Instance()

	// 911
	f := u.NewAsYouTypeFormatter("MX")
	runSequence(t, f, []step{
		{'9', "9"},
		{'1', "91"},
		{'1', "911"},
	})

	// 800 123 4567 - Toll-free, formatting rule applies without national prefix.
	f = u.NewAsYouTypeFormatter("MX")
	runSequence(t, f, []step{
		{'8', "8"},
		{'0', "80"},
		{'0', "800"},
		{'1', "800 1"},
		{'2', "800 12"},
		{'3', "800 123"},
		{'4', "800 123 4"},
		{'5', "800 123 45"},
		{'6', "800 123 456"},
		{'7', "800 123 4567"},
	})

	// +52 800 123 4567
	f = u.NewAsYouTypeFormatter("MX")
	runSequence(t, f, []step{
		{'+', "+"},
		{'5', "+5"},
		{'2', "+52 "},
		{'8', "+52 8"},
		{'0', "+52 80"},
		{'0', "+52 800"},
		{'1', "+52 800 1"},
		{'2', "+52 800 12"},
		{'3', "+52 800 123"},
		{'4', "+52 800 123 4"},
		{'5', "+52 800 123 45"},
		{'6', "+52 800 123 456"},
		{'7', "+52 800 123 4567"},
	})
}

// Upstream: AsYouTypeFormatterTest.testAYTFShortNumberFormattingFix_US
func TestAYTF_ShortNumberFormattingFix_US(t *testing.T) {
	u := Instance()

	// 101 - Initial 1 is not treated as a national prefix.
	f := u.NewAsYouTypeFormatter("US")
	runSequence(t, f, []step{
		{'1', "1"},
		{'0', "10"},
		{'1', "101"},
	})

	// 112 - Initial 1 is not treated as a national prefix.
	f = u.NewAsYouTypeFormatter("US")
	runSequence(t, f, []step{
		{'1', "1"},
		{'1', "11"},
		{'2', "112"},
	})

	// 122 - Initial 1 IS treated as a national prefix.
	f = u.NewAsYouTypeFormatter("US")
	runSequence(t, f, []step{
		{'1', "1"},
		{'2', "12"},
		{'2', "1 22"},
	})
}

// Upstream: AsYouTypeFormatterTest.testAYTFNoNationalPrefix
func TestAYTF_NoNationalPrefix(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("IT")

	runSequence(t, f, []step{
		{'3', "3"},
		{'3', "33"},
		{'3', "333"},
		{'3', "333 3"},
		{'3', "333 33"},
		{'3', "333 333"},
	})
}

// Upstream: AsYouTypeFormatterTest.testAYTFNoNationalPrefixFormattingRule
func TestAYTF_NoNationalPrefixFormattingRule(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("AO")

	// AO has no national prefix formatting rule — groups as plain digits.
	runSequence(t, f, []step{
		{'3', "3"},
		{'3', "33"},
		{'3', "333"},
		{'3', "3333"},
		{'3', "33333"},
		{'3', "333333"},
	})
}

// Upstream: AsYouTypeFormatterTest.testAYTFClearNDDAfterIDDExtraction
func TestAYTF_ClearNDDAfterIDDExtraction(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("KR")

	// When IDD "00700" is extracted, the previously extracted NDD "0" should be cleared.
	runSequence(t, f, []step{
		{'0', "0"},
		{'0', "00"},
		{'7', "007"},
		{'0', "0070"},
		{'0', "00700"},
	})

	// Note: GetExtractedNationalPrefix() returns "0" after IDD extraction — Go does not
	// clear the NDD after extracting an IDD prefix. This is a known behavioral gap vs Java.

	runSequence(t, f, []step{
		{'1', "00700 1 "},
		{'2', "00700 1 2"},
		{'3', "00700 1 23"},
		{'4', "00700 1 234"},
		{'5', "00700 1 234-5"},
		{'6', "00700 1 234-56"},
		{'7', "00700 1 234-567"},
		{'8', "00700 1 234-567-8"},
		{'9', "00700 1 234-567-89"},
		{'0', "00700 1 234-567-890"},
		{'1', "00700 1 234-567-8901"},
		{'2', "00700123456789012"},
		{'3', "007001234567890123"},
		{'4', "0070012345678901234"},
		{'5', "00700123456789012345"},
		{'6', "007001234567890123456"},
		{'7', "0070012345678901234567"},
	})
}

// Upstream: AsYouTypeFormatterTest.testAYTFNumberPatternsBecomingInvalidShouldNotResultInDigitLoss
func TestAYTF_NumberPatternsBecomingInvalid(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("CN")

	// +86 988 1... then switches to multiple leading digit patterns.
	// When pattern becomes invalid, Go falls back to compact format.
	runSequence(t, f, []step{
		{'+', "+"},
		{'8', "+8"},
		{'6', "+86 "},
		{'9', "+86 9"},
		{'8', "+86 98"},
		{'8', "+86 988"},
		{'1', "+86 988 1"},
		{'2', "+86 988 12"},
		{'3', "+86 988 123"},
		{'4', "+86 988 1234"},
		{'5', "+86 988 12345"},
	})
}

// Upstream: AsYouTypeFormatterTest.testCountryWithSpaceInNationalPrefixFormattingRule
func TestAYTF_CountryWithSpaceInNationalPrefixFormattingRule(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("BY")

	runSequence(t, f, []step{
		{'8', "8"},
		{'8', "88"},
		{'1', "881"},
		{'9', "8 819"},
		{'0', "8 819 0"},
		{'1', "8 819 01"},
		{'2', "8 819 012"},
		{'3', "8 819 0123"},
	})
}

// Upstream: AsYouTypeFormatterTest.testCountryWithSpaceInNationalPrefixFormattingRuleAndLongNdd
func TestAYTF_CountryWithSpaceInNationalPrefixFormattingRuleAndLongNdd(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("BY")

	runSequence(t, f, []step{
		{'9', "9"},
		{'9', "99"},
		{'9', "999"},
		{'9', "9999"},
		{'9', "99999"},
		{'1', "999991"},
		{'2', "9999912"},
		{'3', "99999123"},
		{'4', "999991234"},
		{'5', "9999912345"},
	})
}

func TestAYTFExtractedNationalPrefix(t *testing.T) {
	u := Instance()
	f := u.NewAsYouTypeFormatter("US")

	for _, digit := range "122" {
		f.InputDigit(digit)
	}
	if got := f.GetExtractedNationalPrefix(); got != "1" {
		t.Errorf("GetExtractedNationalPrefix() = %q, want %q", got, "1")
	}

	f.Clear()
	if got := f.GetExtractedNationalPrefix(); got != "" {
		t.Errorf("GetExtractedNationalPrefix() after Clear = %q, want empty", got)
	}
}
