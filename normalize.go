package phonesafe

import "strings"

// alphaToDigit maps alphabetic characters to their phone keypad digits.
var alphaToDigit = map[rune]rune{
	'A': '2', 'B': '2', 'C': '2',
	'D': '3', 'E': '3', 'F': '3',
	'G': '4', 'H': '4', 'I': '4',
	'J': '5', 'K': '5', 'L': '5',
	'M': '6', 'N': '6', 'O': '6',
	'P': '7', 'Q': '7', 'R': '7', 'S': '7',
	'T': '8', 'U': '8', 'V': '8',
	'W': '9', 'X': '9', 'Y': '9', 'Z': '9',
	'a': '2', 'b': '2', 'c': '2',
	'd': '3', 'e': '3', 'f': '3',
	'g': '4', 'h': '4', 'i': '4',
	'j': '5', 'k': '5', 'l': '5',
	'm': '6', 'n': '6', 'o': '6',
	'p': '7', 'q': '7', 'r': '7', 's': '7',
	't': '8', 'u': '8', 'v': '8',
	'w': '9', 'x': '9', 'y': '9', 'z': '9',
}

// NormalizeDigitsOnly strips all non-digit characters and converts
// Unicode decimal digits (Arabic-Indic, Eastern Arabic-Indic, fullwidth) to ASCII.
func NormalizeDigitsOnly(number string) string {
	var b strings.Builder
	b.Grow(len(number))
	for _, r := range number {
		if d, ok := unicodeDigitToASCII(r); ok {
			b.WriteByte(d)
		}
	}
	return b.String()
}

// NormalizeDiallableCharsOnly removes all characters that are not valid
// diallable characters (digits, +, *, #).
func NormalizeDiallableCharsOnly(number string) string {
	var b strings.Builder
	b.Grow(len(number))
	for _, r := range number {
		if d, ok := unicodeDigitToASCII(r); ok {
			b.WriteByte(d)
		} else {
			switch r {
			case '+', '＋':
				b.WriteByte('+')
			case '*':
				b.WriteByte('*')
			case '#':
				b.WriteByte('#')
			}
		}
	}
	return b.String()
}

// ConvertAlphaCharactersInNumber converts alpha characters in a phone number
// to their corresponding keypad digits, leaving other characters unchanged.
func ConvertAlphaCharactersInNumber(number string) string {
	var b strings.Builder
	b.Grow(len(number))
	for _, r := range number {
		if digit, ok := alphaToDigit[r]; ok {
			b.WriteRune(digit)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// normalize converts a phone number string by:
// 1. If 3+ alpha characters are present, converting them to keypad digits.
// 2. Stripping non-digits and normalizing Unicode digits to ASCII.
func normalize(number string) string {
	alphaCount := 0
	for _, r := range number {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			alphaCount++
			if alphaCount >= 3 {
				break
			}
		}
	}

	if alphaCount >= 3 {
		// Convert alpha chars to digits, then strip non-digits
		var b strings.Builder
		b.Grow(len(number))
		for _, r := range number {
			if digit, ok := alphaToDigit[r]; ok {
				b.WriteRune(digit)
			} else if d, ok := unicodeDigitToASCII(r); ok {
				b.WriteByte(d)
			}
		}
		return b.String()
	}

	return NormalizeDigitsOnly(number)
}

// unicodeDigitToASCII converts a Unicode decimal digit rune to its ASCII byte
// equivalent. Returns the byte and true if the rune is a decimal digit, or
// 0 and false otherwise.
func unicodeDigitToASCII(r rune) (byte, bool) {
	switch {
	case r >= '0' && r <= '9':
		return byte(r), true
	case r >= '٠' && r <= '٩': // Arabic-Indic
		return byte(r-'٠') + '0', true
	case r >= '۰' && r <= '۹': // Extended Arabic-Indic
		return byte(r-'۰') + '0', true
	case r >= '０' && r <= '９': // Fullwidth
		return byte(r-'０') + '0', true
	case r >= '०' && r <= '९': // Devanagari
		return byte(r-'०') + '0', true
	case r >= '০' && r <= '৯': // Bengali
		return byte(r-'০') + '0', true
	default:
		return 0, false
	}
}
