// Command phonesafe-gen converts upstream libphonenumber metadata XML files
// into generated Go source files for the go-phonesafe library.
//
// Usage:
//
//	phonesafe-gen [flags]
//
// Flags:
//
//	-input string          Path to libphonenumber resources/ directory (required)
//	-output string         Path to output directory for generated files (required unless -validate-only)
//	-upstream-version string  Upstream libphonenumber version tag (e.g., "v8.13.50") (required unless -validate-only)
//	-languages string      Comma-separated list of geocoding languages to include (default: all)
//	-lite                  Exclude example numbers from metadata (smaller output)
//	-validate-only         Parse and validate without emitting files
//	-verbose               Print progress information
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var verbose bool

func logVerbose(format string, args ...any) {
	if verbose {
		fmt.Printf(format+"\n", args...)
	}
}

func main() {
	inputDir := flag.String("input", "", "Path to libphonenumber resources/ directory (required)")
	outputDir := flag.String("output", "", "Path to output directory for generated files (required unless -validate-only)")
	upstreamVersion := flag.String("upstream-version", "", "Upstream libphonenumber version tag (e.g., \"v8.13.50\") (required unless -validate-only)")
	languages := flag.String("languages", "", "Comma-separated list of geocoding languages to include (default: all)")
	lite := flag.Bool("lite", false, "Exclude example numbers from metadata (smaller output)")
	validateOnly := flag.Bool("validate-only", false, "Parse and validate without emitting files")
	flag.BoolVar(&verbose, "verbose", false, "Print progress information")
	flag.Parse()

	if *inputDir == "" {
		fmt.Fprintf(os.Stderr, "Error: -input is required\n")
		flag.Usage()
		os.Exit(1)
	}
	if !*validateOnly && (*outputDir == "" || *upstreamVersion == "") {
		fmt.Fprintf(os.Stderr, "Error: -output and -upstream-version are required unless -validate-only is set\n")
		flag.Usage()
		os.Exit(1)
	}

	// Resolve absolute paths.
	absInput, err := filepath.Abs(*inputDir)
	if err != nil {
		log.Fatalf("resolving input path: %v", err)
	}

	// Verify input files exist.
	requiredFiles := []string{
		"PhoneNumberMetadata.xml",
		"ShortNumberMetadata.xml",
		"PhoneNumberAlternateFormats.xml",
	}
	for _, f := range requiredFiles {
		path := filepath.Join(absInput, f)
		if _, err := os.Stat(path); err != nil {
			log.Fatalf("required input file not found: %s", path)
		}
	}

	// Verify geocoding and timezone dirs.
	geocodingDir := filepath.Join(absInput, "geocoding")
	if _, err := os.Stat(geocodingDir); err != nil {
		log.Fatalf("geocoding directory not found: %s", geocodingDir)
	}
	timezoneFile := filepath.Join(absInput, "timezones", "map_data.txt")
	if _, err := os.Stat(timezoneFile); err != nil {
		log.Fatalf("timezone data file not found: %s", timezoneFile)
	}

	// Parse language filter.
	var langFilter map[string]struct{}
	if *languages != "" {
		langFilter = make(map[string]struct{})
		for _, l := range strings.Split(*languages, ",") {
			l = strings.TrimSpace(l)
			if l != "" {
				langFilter[l] = struct{}{}
			}
		}
	}

	// 1. Parse phone number metadata.
	logVerbose("Parsing PhoneNumberMetadata.xml...")
	phoneMetadata, err := parsePhoneMetadataXML(filepath.Join(absInput, "PhoneNumberMetadata.xml"))
	if err != nil {
		log.Fatalf("parsing PhoneNumberMetadata.xml: %v", err)
	}
	logVerbose("  Parsed %d territories", len(phoneMetadata))

	// 2. Parse short number metadata.
	logVerbose("Parsing ShortNumberMetadata.xml...")
	shortMetadata, err := parseShortMetadataXML(filepath.Join(absInput, "ShortNumberMetadata.xml"))
	if err != nil {
		log.Fatalf("parsing ShortNumberMetadata.xml: %v", err)
	}
	logVerbose("  Parsed %d short number territories", len(shortMetadata))

	// 3. Parse alternate formats.
	logVerbose("Parsing PhoneNumberAlternateFormats.xml...")
	altFormats, err := parseAlternateFormatsXML(filepath.Join(absInput, "PhoneNumberAlternateFormats.xml"))
	if err != nil {
		log.Fatalf("parsing PhoneNumberAlternateFormats.xml: %v", err)
	}
	logVerbose("  Parsed %d country alternate format sets", len(altFormats))

	// 4. Parse geocoding data.
	logVerbose("Parsing geocoding data...")
	geoData, err := parseGeocodingDir(geocodingDir)
	if err != nil {
		log.Fatalf("parsing geocoding data: %v", err)
	}
	logVerbose("  Parsed %d languages", len(geoData))

	// Apply language filter.
	if langFilter != nil {
		for lang := range geoData {
			if _, ok := langFilter[lang]; !ok {
				delete(geoData, lang)
			}
		}
		logVerbose("  Filtered to %d languages", len(geoData))
	}

	// 5. Parse timezone data.
	logVerbose("Parsing timezone data...")
	tzData, err := parseTimezoneFile(timezoneFile)
	if err != nil {
		log.Fatalf("parsing timezone data: %v", err)
	}
	logVerbose("  Parsed %d timezone entries", len(tzData))

	// If -lite, strip example numbers.
	if *lite {
		stripExampleNumbers(phoneMetadata)
		stripShortExampleNumbers(shortMetadata)
		logVerbose("  Stripped example numbers (-lite)")
	}

	// Validate-only mode: exit after successful parse.
	if *validateOnly {
		fmt.Println("Validation passed.")
		return
	}

	// Ensure output directories exist.
	absOutput, err := filepath.Abs(*outputDir)
	if err != nil {
		log.Fatalf("resolving output path: %v", err)
	}
	metadataOutDir := filepath.Join(absOutput, "internal", "metadata")
	dataOutDir := filepath.Join(absOutput, "internal", "data")
	if err := os.MkdirAll(metadataOutDir, 0o755); err != nil {
		log.Fatalf("creating metadata output dir: %v", err)
	}
	if err := os.MkdirAll(dataOutDir, 0o755); err != nil {
		log.Fatalf("creating data output dir: %v", err)
	}

	// Emit generated files.
	logVerbose("Emitting metadata_gen.go...")
	if err := emitMetadata(phoneMetadata, metadataOutDir, *upstreamVersion); err != nil {
		log.Fatalf("emitting metadata: %v", err)
	}

	logVerbose("Emitting shortmetadata_gen.go...")
	if err := emitShortMetadata(shortMetadata, metadataOutDir, *upstreamVersion); err != nil {
		log.Fatalf("emitting short metadata: %v", err)
	}

	logVerbose("Emitting alternateformats_gen.go...")
	if err := emitAlternateFormats(altFormats, metadataOutDir, *upstreamVersion); err != nil {
		log.Fatalf("emitting alternate formats: %v", err)
	}

	logVerbose("Emitting geocoding_gen.go...")
	if err := emitGeocoding(geoData, dataOutDir, *upstreamVersion); err != nil {
		log.Fatalf("emitting geocoding data: %v", err)
	}

	logVerbose("Emitting timezones_gen.go...")
	if err := emitTimezones(tzData, dataOutDir, *upstreamVersion); err != nil {
		log.Fatalf("emitting timezone data: %v", err)
	}

	logVerbose("Emitting version_gen.go...")
	if err := emitVersion(*upstreamVersion, metadataOutDir); err != nil {
		log.Fatalf("emitting version: %v", err)
	}

	logVerbose("Code generation complete.")
}

// stripExampleNumbers removes ExampleNumber from all phone metadata descriptors.
func stripExampleNumbers(territories []parsedTerritory) {
	for i := range territories {
		t := &territories[i]
		t.GeneralDesc.ExampleNumber = ""
		t.FixedLine.ExampleNumber = ""
		t.Mobile.ExampleNumber = ""
		t.TollFree.ExampleNumber = ""
		t.PremiumRate.ExampleNumber = ""
		t.SharedCost.ExampleNumber = ""
		t.PersonalNumber.ExampleNumber = ""
		t.VOIP.ExampleNumber = ""
		t.Pager.ExampleNumber = ""
		t.UAN.ExampleNumber = ""
		t.Voicemail.ExampleNumber = ""
		t.NoInternationalDialling.ExampleNumber = ""
	}
}

// stripShortExampleNumbers removes ExampleNumber from all short number metadata.
func stripShortExampleNumbers(territories []parsedShortTerritory) {
	for i := range territories {
		t := &territories[i]
		t.GeneralDesc.ExampleNumber = ""
		t.TollFree.ExampleNumber = ""
		t.StandardRate.ExampleNumber = ""
		t.PremiumRate.ExampleNumber = ""
		t.CarrierSpecific.ExampleNumber = ""
		t.Emergency.ExampleNumber = ""
		t.ExpandedEmergency.ExampleNumber = ""
		t.SMSServices.ExampleNumber = ""
	}
}
