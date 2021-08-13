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

const (
	max_bills = 30
)

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
	all := flag.Bool("all", false, "processes all bills-- otherwise process a sample")

	// allow user to pass billnumbers as argument
	var billList BillList
	var sampleSize int
	flag.Var(&billList, "billnumbers", "comma-separated list of billnumbers")
	flag.Var(&billList, "b", "comma-separated list of billnumbers")
	flag.IntVar(&sampleSize, "samplesize", 0, "number of sections to sample in large bill")
	flag.Parse()

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
	//bills.PrintESInfo()
	//bills.SampleQuery()

	var billNumbers []string
	if *all {
		billNumbers = bills.GetAllBillNumbers()
	} else if len(billList) > 0 {
		billNumbers = billList
	} else {
		billNumbers = bills.GetSampleBillNumbers()
	}
	for _, billnumber := range billNumbers {
		similaritySectionsByBillNumber := bills.GetSimilaritySectionsByBillNumber(billnumber, sampleSize)
		similarBillsDict := bills.GetSimilarBillsDict(similaritySectionsByBillNumber, max_bills)
		log.Info().Msgf("Similar Bills Dict: %v", similarBillsDict)
		//similarBillMapBySection := bills.SimilarSectionsItemsToBillMap(similaritySectionsByBillNumber)
		//bills := bills.GetSimilarBills(similarBillMapBySection)
		//log.Info().Msgf("Similar Bills: %v", bills)
		//TODO Select top bills based on score
		//Find how many sections and how many matches
	}
}
