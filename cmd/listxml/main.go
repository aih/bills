package main

import (
	"flag"

	"github.com/aih/bills"
)

func main() {
	flagUsage := "Absolute path to the 'congress' directory, where govinfo data is downloaded"
	flagValue := string(bills.PathToCongressDataDir)
	var pathToCongressDataDir string
	flag.StringVar(&pathToCongressDataDir, "congressPath", flagValue, flagUsage)
	flag.StringVar(&pathToCongressDataDir, "c", string(bills.PathToCongressDataDir), flagUsage+" (shorthand)")
	flag.Parse()
	bills.ListDocumentXMLFiles(pathToCongressDataDir)
}
