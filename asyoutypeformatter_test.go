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
