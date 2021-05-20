package main

import (
	"flag"
	"os"

	"github.com/aih/bills"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	debug := flag.Bool("debug", false, "sets log level to debug")

	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	flagUsage := "Absolute path to the 'congress' directory, where govinfo data is downloaded"
	flagValue := string(bills.PathToCongressDataDir)
	var pathToCongressDataDir string
	flag.StringVar(&pathToCongressDataDir, "congressPath", flagValue, flagUsage)
	flag.StringVar(&pathToCongressDataDir, "c", string(bills.PathToCongressDataDir), flagUsage+" (shorthand)")
	flag.Parse()

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	log.Debug().Msg("Log level set to Debug")
	bills.ListDocumentXMLFiles(pathToCongressDataDir)
}
