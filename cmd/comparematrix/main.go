package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aih/bills"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

/*
func main() {
	bills.CompareSamples()
}
*/

// BillList is a string slice
type BillList []string

func (bl *BillList) String() string {
	return fmt.Sprintln(*bl)
}

// Set string value in MyList
func (bl *BillList) Set(s string) error {
	*bl = strings.Split(s, ",")
	return nil
}

func main() {

	debug := flag.Bool("debug", false, "sets log level to debug")

	flag.Parse()

	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Debug().Msg("Log level set to Debug")

	flagPathUsage := "Absolute path to the parent directory for 'congress' and json metadata files"
	flagPathValue := string(bills.ParentPathDefault)
	var parentPath string
	flag.StringVar(&parentPath, "parentPath", flagPathValue, flagPathUsage)
	flag.StringVar(&parentPath, "p", flagPathValue, flagPathUsage+" (shorthand)")

	var billList BillList
	flag.Var(&billList, "billnumbers", "comma-separated list of billnumbers")
	flag.Var(&billList, "b", "comma-separated list of billnumbers")
	flag.Parse()

	bills.CompareBills(parentPath, billList)
}
