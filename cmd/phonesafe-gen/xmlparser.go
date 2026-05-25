package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// parsedTerritory holds the parsed metadata for a single region.
type parsedTerritory struct {
	ID                           string
	CountryCode                  int32
	InternationalPrefix          string
	PreferredInternationalPrefix string
	NationalPrefix               string
	PreferredExtnPrefix          string
	NationalPrefixForParsing     string
	NationalPrefixTransformRule  string
	MainCountryForCode           bool
	LeadingDigits                string
	MobileNumberPortable         bool

	GeneralDesc             parsedDesc
	FixedLine               parsedDesc
	Mobile                  parsedDesc
	TollFree                parsedDesc
	PremiumRate             parsedDesc
	SharedCost              parsedDesc
	PersonalNumber          parsedDesc
	VOIP                    parsedDesc
	Pager                   parsedDesc
	UAN                     parsedDesc
	Voicemail               parsedDesc
	NoInternationalDialling parsedDesc

	NumberFormats     []parsedFormat
	IntlNumberFormats []parsedFormat
}

// parsedDesc holds a phone number type descriptor.
type parsedDesc struct {
	NationalNumberPattern string
	PossibleLengths       []int32
	PossibleLengthsLocal  []int32
	ExampleNumber         string
}

// parsedFormat holds a number formatting rule.
type parsedFormat struct {
	Pattern                          string
	Format                           string
	LeadingDigitsPatterns            []string
	NationalPrefixFormattingRule     string
	NationalPrefixOptional           bool
	DomesticCarrierCodeFormattingRule string
	IntlFormat                       string
}

// parsedShortTerritory holds short number metadata for a region.
type parsedShortTerritory struct {
	ID                string
	GeneralDesc       parsedDesc
	TollFree          parsedDesc
	StandardRate      parsedDesc
	PremiumRate       parsedDesc
	CarrierSpecific   parsedDesc
	Emergency         parsedDesc
	ExpandedEmergency parsedDesc
	SMSServices       parsedDesc
}

// parsedAltFormats holds alternate formatting rules for a country code.
type parsedAltFormats struct {
	CountryCode int32
	Formats     []parsedFormat
}

// geoEntry maps a prefix to a geographic description.
type geoEntry struct {
	Prefix      int32
	Description string
}

// tzEntry maps a prefix to timezone IDs.
type tzEntry struct {
	Prefix    int32
	TimeZones []string
}

// --- XML types for decoding ---

type xmlPhoneNumberMetadata struct {
	XMLName     xml.Name         `xml:"phoneNumberMetadata"`
	Territories xmlTerritories   `xml:"territories"`
}

type xmlTerritories struct {
	Territory []xmlTerritory `xml:"territory"`
}

type xmlTerritory struct {
	ID                           string `xml:"id,attr"`
	CountryCode                  string `xml:"countryCode,attr"`
	InternationalPrefix          string `xml:"internationalPrefix,attr"`
	PreferredInternationalPrefix string `xml:"preferredInternationalPrefix,attr"`
	NationalPrefix               string `xml:"nationalPrefix,attr"`
	PreferredExtnPrefix          string `xml:"preferredExtnPrefix,attr"`
	NationalPrefixForParsing     string `xml:"nationalPrefixForParsing,attr"`
	NationalPrefixTransformRule  string `xml:"nationalPrefixTransformRule,attr"`
	MainCountryForCode           string `xml:"mainCountryForCode,attr"`
	LeadingDigits                string `xml:"leadingDigits,attr"`
	MobileNumberPortableRegion   string `xml:"mobileNumberPortableRegion,attr"`

	AvailableFormats *xmlAvailableFormats `xml:"availableFormats"`

	GeneralDesc             *xmlDesc `xml:"generalDesc"`
	FixedLine               *xmlDesc `xml:"fixedLine"`
	Mobile                  *xmlDesc `xml:"mobile"`
	TollFree                *xmlDesc `xml:"tollFree"`
	PremiumRate             *xmlDesc `xml:"premiumRate"`
	SharedCost              *xmlDesc `xml:"sharedCost"`
	PersonalNumber          *xmlDesc `xml:"personalNumber"`
	VOIP                    *xmlDesc `xml:"voip"`
	Pager                   *xmlDesc `xml:"pager"`
	UAN                     *xmlDesc `xml:"uan"`
	Voicemail               *xmlDesc `xml:"voicemail"`
	NoInternationalDialling *xmlDesc `xml:"noInternationalDialling"`

	// Short number specific elements
	ShortCode         *xmlDesc `xml:"shortCode"`
	StandardRate      *xmlDesc `xml:"standardRate"`
	CarrierSpecific   *xmlDesc `xml:"carrierSpecific"`
	Emergency         *xmlDesc `xml:"emergency"`
	ExpandedEmergency *xmlDesc `xml:"expandedEmergency"`
	SMSServices       *xmlDesc `xml:"smsServices"`
}

type xmlAvailableFormats struct {
	NumberFormat []xmlNumberFormat `xml:"numberFormat"`
}

type xmlNumberFormat struct {
	Pattern                          string   `xml:"pattern,attr"`
	NationalPrefixFormattingRule     string   `xml:"nationalPrefixFormattingRule,attr"`
	NationalPrefixOptionalWhenFormatting string `xml:"nationalPrefixOptionalWhenFormatting,attr"`
	CarrierCodeFormattingRule        string   `xml:"carrierCodeFormattingRule,attr"`
	LeadingDigits                    []string `xml:"leadingDigits"`
	Format                           string   `xml:"format"`
	IntlFormat                       string   `xml:"intlFormat"`
}

type xmlDesc struct {
	NationalNumberPattern string          `xml:"nationalNumberPattern"`
	PossibleLengths       *xmlPossibleLen `xml:"possibleLengths"`
	ExampleNumber         string          `xml:"exampleNumber"`
}

type xmlPossibleLen struct {
	National  string `xml:"national,attr"`
	LocalOnly string `xml:"localOnly,attr"`
}

// --- Parsing functions ---

func parsePhoneMetadataXML(path string) ([]parsedTerritory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var doc xmlPhoneNumberMetadata
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("unmarshaling XML: %w", err)
	}

	var territories []parsedTerritory
	for _, xt := range doc.Territories.Territory {
		t := convertTerritory(xt)
		territories = append(territories, t)
	}
	return territories, nil
}

func parseShortMetadataXML(path string) ([]parsedShortTerritory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var doc xmlPhoneNumberMetadata
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("unmarshaling XML: %w", err)
	}

	var territories []parsedShortTerritory
	for _, xt := range doc.Territories.Territory {
		st := parsedShortTerritory{
			ID: xt.ID,
		}
		if xt.ShortCode != nil {
			st.GeneralDesc = convertDesc(xt.ShortCode)
		} else if xt.GeneralDesc != nil {
			st.GeneralDesc = convertDesc(xt.GeneralDesc)
		}
		if xt.TollFree != nil {
			st.TollFree = convertDesc(xt.TollFree)
		}
		if xt.StandardRate != nil {
			st.StandardRate = convertDesc(xt.StandardRate)
		}
		if xt.PremiumRate != nil {
			st.PremiumRate = convertDesc(xt.PremiumRate)
		}
		if xt.CarrierSpecific != nil {
			st.CarrierSpecific = convertDesc(xt.CarrierSpecific)
		}
		if xt.Emergency != nil {
			st.Emergency = convertDesc(xt.Emergency)
		}
		if xt.ExpandedEmergency != nil {
			st.ExpandedEmergency = convertDesc(xt.ExpandedEmergency)
		}
		if xt.SMSServices != nil {
			st.SMSServices = convertDesc(xt.SMSServices)
		}
		territories = append(territories, st)
	}
	return territories, nil
}

func parseAlternateFormatsXML(path string) ([]parsedAltFormats, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var doc xmlPhoneNumberMetadata
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("unmarshaling XML: %w", err)
	}

	var result []parsedAltFormats
	for _, xt := range doc.Territories.Territory {
		cc, _ := strconv.ParseInt(xt.CountryCode, 10, 32)
		af := parsedAltFormats{
			CountryCode: int32(cc),
		}
		if xt.AvailableFormats != nil {
			for _, xf := range xt.AvailableFormats.NumberFormat {
				af.Formats = append(af.Formats, convertFormat(xf))
			}
		}
		result = append(result, af)
	}
	return result, nil
}

func convertTerritory(xt xmlTerritory) parsedTerritory {
	cc, _ := strconv.ParseInt(xt.CountryCode, 10, 32)
	t := parsedTerritory{
		ID:                           xt.ID,
		CountryCode:                  int32(cc),
		InternationalPrefix:          stripPatternWS(xt.InternationalPrefix),
		PreferredInternationalPrefix: xt.PreferredInternationalPrefix,
		NationalPrefix:               xt.NationalPrefix,
		PreferredExtnPrefix:          xt.PreferredExtnPrefix,
		NationalPrefixForParsing:     stripPatternWS(xt.NationalPrefixForParsing),
		NationalPrefixTransformRule:  xt.NationalPrefixTransformRule,
		MainCountryForCode:           xt.MainCountryForCode == "true",
		LeadingDigits:                stripPatternWS(xt.LeadingDigits),
		MobileNumberPortable:         xt.MobileNumberPortableRegion == "true",
	}

	if xt.GeneralDesc != nil {
		t.GeneralDesc = convertDesc(xt.GeneralDesc)
	}
	if xt.FixedLine != nil {
		t.FixedLine = convertDesc(xt.FixedLine)
	}
	if xt.Mobile != nil {
		t.Mobile = convertDesc(xt.Mobile)
	}
	if xt.TollFree != nil {
		t.TollFree = convertDesc(xt.TollFree)
	}
	if xt.PremiumRate != nil {
		t.PremiumRate = convertDesc(xt.PremiumRate)
	}
	if xt.SharedCost != nil {
		t.SharedCost = convertDesc(xt.SharedCost)
	}
	if xt.PersonalNumber != nil {
		t.PersonalNumber = convertDesc(xt.PersonalNumber)
	}
	if xt.VOIP != nil {
		t.VOIP = convertDesc(xt.VOIP)
	}
	if xt.Pager != nil {
		t.Pager = convertDesc(xt.Pager)
	}
	if xt.UAN != nil {
		t.UAN = convertDesc(xt.UAN)
	}
	if xt.Voicemail != nil {
		t.Voicemail = convertDesc(xt.Voicemail)
	}
	if xt.NoInternationalDialling != nil {
		t.NoInternationalDialling = convertDesc(xt.NoInternationalDialling)
	}

	if xt.AvailableFormats != nil {
		for _, xf := range xt.AvailableFormats.NumberFormat {
			pf := convertFormat(xf)
			t.NumberFormats = append(t.NumberFormats, pf)
		}
		// Build IntlNumberFormats: skip formats with IntlFormat="NA",
		// use IntlFormat if provided, otherwise copy the national format.
		for _, xf := range xt.AvailableFormats.NumberFormat {
			if xf.IntlFormat == "NA" {
				continue
			}
			pf := convertFormat(xf)
			if xf.IntlFormat != "" {
				pf.Format = strings.TrimSpace(xf.IntlFormat)
			}
			// IntlFormat field is not needed in the intl copy
			pf.IntlFormat = ""
			t.IntlNumberFormats = append(t.IntlNumberFormats, pf)
		}
	}

	return t
}

func convertDesc(xd *xmlDesc) parsedDesc {
	d := parsedDesc{
		NationalNumberPattern: stripPatternWS(xd.NationalNumberPattern),
		ExampleNumber:         strings.TrimSpace(xd.ExampleNumber),
	}
	if xd.PossibleLengths != nil {
		d.PossibleLengths = parseLengths(xd.PossibleLengths.National)
		d.PossibleLengthsLocal = parseLengths(xd.PossibleLengths.LocalOnly)
	}
	return d
}

func convertFormat(xf xmlNumberFormat) parsedFormat {
	pf := parsedFormat{
		Pattern:                          xf.Pattern,
		Format:                           strings.TrimSpace(xf.Format),
		NationalPrefixFormattingRule:     xf.NationalPrefixFormattingRule,
		NationalPrefixOptional:           xf.NationalPrefixOptionalWhenFormatting == "true",
		DomesticCarrierCodeFormattingRule: xf.CarrierCodeFormattingRule,
		IntlFormat:                       strings.TrimSpace(xf.IntlFormat),
	}
	for _, ld := range xf.LeadingDigits {
		pf.LeadingDigitsPatterns = append(pf.LeadingDigitsPatterns, stripPatternWS(ld))
	}
	return pf
}

// stripPatternWS removes all whitespace from a regex pattern.
// libphonenumber XML uses whitespace for readability in patterns.
func stripPatternWS(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\n', '\r':
			return -1
		default:
			return r
		}
	}, s)
}

// parseLengths parses possible lengths from the XML attribute format.
// Handles: "5", "5,7,9", "[3-6]", "[3-6],8,10", "-1" (meaning not set).
func parseLengths(s string) []int32 {
	s = strings.TrimSpace(s)
	if s == "" || s == "-1" {
		return nil
	}

	var result []int32
	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			// Range like [3-6]
			inner := part[1 : len(part)-1]
			rangeParts := strings.Split(inner, "-")
			if len(rangeParts) == 2 {
				lo, _ := strconv.Atoi(rangeParts[0])
				hi, _ := strconv.Atoi(rangeParts[1])
				for i := lo; i <= hi; i++ {
					result = append(result, int32(i))
				}
			}
		} else {
			v, err := strconv.Atoi(part)
			if err == nil && v > 0 {
				result = append(result, int32(v))
			}
		}
	}
	return result
}

// --- Text data parsers ---

func parseGeocodingDir(dir string) (map[string][]geoEntry, error) {
	result := make(map[string][]geoEntry)

	// Walk language directories.
	langDirs, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading geocoding dir: %w", err)
	}

	for _, langDir := range langDirs {
		if !langDir.IsDir() {
			continue
		}
		lang := langDir.Name()
		langPath := filepath.Join(dir, lang)

		files, err := os.ReadDir(langPath)
		if err != nil {
			return nil, fmt.Errorf("reading lang dir %s: %w", lang, err)
		}

		var entries []geoEntry
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".txt") {
				continue
			}
			fpath := filepath.Join(langPath, f.Name())
			fileEntries, err := parseGeocodingFile(fpath)
			if err != nil {
				return nil, fmt.Errorf("parsing %s: %w", fpath, err)
			}
			entries = append(entries, fileEntries...)
		}

		// Sort by prefix.
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Prefix < entries[j].Prefix
		})
		result[lang] = entries
	}

	return result, nil
}

func parseGeocodingFile(path string) ([]geoEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []geoEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}
		prefix, err := strconv.ParseInt(parts[0], 10, 32)
		if err != nil {
			continue
		}
		entries = append(entries, geoEntry{
			Prefix:      int32(prefix),
			Description: parts[1],
		})
	}
	return entries, scanner.Err()
}

func parseTimezoneFile(path string) ([]tzEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening timezone file: %w", err)
	}
	defer f.Close()

	var entries []tzEntry
	scanner := bufio.NewScanner(f)
	// Some timezone lines are very long.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}
		prefix, err := strconv.ParseInt(parts[0], 10, 32)
		if err != nil {
			continue
		}
		tzs := strings.Split(parts[1], "&")
		entries = append(entries, tzEntry{
			Prefix:    int32(prefix),
			TimeZones: tzs,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning timezone file: %w", err)
	}

	// Sort by prefix.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Prefix < entries[j].Prefix
	})

	return entries, nil
}
