// Command phonesafe-gen converts upstream libphonenumber metadata XML files
// into generated Go source files for the go-phonesafe library.
//
// Usage:
//
//	phonesafe-gen -input <path-to-resources/> -output <path-to-project-root/>
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	inputDir := flag.String("input", "", "Path to libphonenumber resources/ directory")
	outputDir := flag.String("output", "", "Path to go-phonesafe project root")
	flag.Parse()

	if *inputDir == "" || *outputDir == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Resolve absolute paths.
	absInput, err := filepath.Abs(*inputDir)
	if err != nil {
		log.Fatalf("resolving input path: %v", err)
	}
	absOutput, err := filepath.Abs(*outputDir)
	if err != nil {
		log.Fatalf("resolving output path: %v", err)
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

	// Ensure output directories exist.
	metadataOutDir := filepath.Join(absOutput, "internal", "metadata")
	dataOutDir := filepath.Join(absOutput, "internal", "data")
	if err := os.MkdirAll(metadataOutDir, 0o755); err != nil {
		log.Fatalf("creating metadata output dir: %v", err)
	}
	if err := os.MkdirAll(dataOutDir, 0o755); err != nil {
		log.Fatalf("creating data output dir: %v", err)
	}

	// 1. Parse and emit phone number metadata.
	fmt.Println("Parsing PhoneNumberMetadata.xml...")
	phoneMetadata, err := parsePhoneMetadataXML(filepath.Join(absInput, "PhoneNumberMetadata.xml"))
	if err != nil {
		log.Fatalf("parsing PhoneNumberMetadata.xml: %v", err)
	}
	fmt.Printf("  Parsed %d territories\n", len(phoneMetadata))

	fmt.Println("Emitting metadata_gen.go...")
	if err := emitMetadata(phoneMetadata, metadataOutDir); err != nil {
		log.Fatalf("emitting metadata: %v", err)
	}

	// 2. Parse and emit short number metadata.
	fmt.Println("Parsing ShortNumberMetadata.xml...")
	shortMetadata, err := parseShortMetadataXML(filepath.Join(absInput, "ShortNumberMetadata.xml"))
	if err != nil {
		log.Fatalf("parsing ShortNumberMetadata.xml: %v", err)
	}
	fmt.Printf("  Parsed %d short number territories\n", len(shortMetadata))

	fmt.Println("Emitting shortmetadata_gen.go...")
	if err := emitShortMetadata(shortMetadata, metadataOutDir); err != nil {
		log.Fatalf("emitting short metadata: %v", err)
	}

	// 3. Parse and emit alternate formats.
	fmt.Println("Parsing PhoneNumberAlternateFormats.xml...")
	altFormats, err := parseAlternateFormatsXML(filepath.Join(absInput, "PhoneNumberAlternateFormats.xml"))
	if err != nil {
		log.Fatalf("parsing PhoneNumberAlternateFormats.xml: %v", err)
	}
	fmt.Printf("  Parsed %d country alternate format sets\n", len(altFormats))

	fmt.Println("Emitting alternateformats_gen.go...")
	if err := emitAlternateFormats(altFormats, metadataOutDir); err != nil {
		log.Fatalf("emitting alternate formats: %v", err)
	}

	// 4. Parse and emit geocoding data.
	fmt.Println("Parsing geocoding data...")
	geoData, err := parseGeocodingDir(geocodingDir)
	if err != nil {
		log.Fatalf("parsing geocoding data: %v", err)
	}
	fmt.Printf("  Parsed %d languages\n", len(geoData))

	fmt.Println("Emitting geocoding_gen.go...")
	if err := emitGeocoding(geoData, dataOutDir); err != nil {
		log.Fatalf("emitting geocoding data: %v", err)
	}

	// 5. Parse and emit timezone data.
	fmt.Println("Parsing timezone data...")
	tzData, err := parseTimezoneFile(timezoneFile)
	if err != nil {
		log.Fatalf("parsing timezone data: %v", err)
	}
	fmt.Printf("  Parsed %d timezone entries\n", len(tzData))

	fmt.Println("Emitting timezones_gen.go...")
	if err := emitTimezones(tzData, dataOutDir); err != nil {
		log.Fatalf("emitting timezone data: %v", err)
	}

	fmt.Println("Code generation complete.")
}
