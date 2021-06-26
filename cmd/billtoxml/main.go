package main

import (
	"flag"
	"os"
	"path"

	"github.com/aih/bills"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var sampleFilePath = path.Join("..", "..", "samples", "BILLS-116hr1500eh.xml")

func main() {
	debug := flag.Bool("debug", false, "sets log level to debug")

	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	flag.Parse()

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	log.Debug().Msg("Log level set to Debug")
	parsedBill := bills.ParseBill(sampleFilePath)
	log.Debug().Msgf("Parsed bill, first section: %s", parsedBill.Sections[1].OutputXML(true))

}
