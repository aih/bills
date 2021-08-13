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
	max_bills = 15
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
	var parentPath string
	flag.Var(&billList, "billnumbers", "comma-separated list of billnumbers")
	flag.Var(&billList, "b", "comma-separated list of billnumbers")
	flag.IntVar(&sampleSize, "samplesize", 0, "number of sections to sample in large bill")
	flagPathUsage := "Absolute path to the parent directory for 'congress' and json metadata files"
	flagPathValue := string(bills.ParentPathDefault)
	flag.StringVar(&parentPath, "parentPath", flagPathValue, flagPathUsage)
	flag.StringVar(&parentPath, "p", flagPathValue, flagPathUsage+" (shorthand)")

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
		// This is the equivalent of es_similarity in BillMap
		log.Info().Msgf("Get versions of: %s", billnumber)
		r := bills.GetBill_ES(billnumber)
		log.Info().Msgf("Number of versions of: %s, %d", billnumber, len(r["hits"].(map[string]interface{})["hits"].([]interface{})))
		latestBillItem, err := bills.GetLatestBill(r)
		if err != nil {
			log.Error().Msgf("Error getting latest bill: '%v'", err)
		}
		similaritySectionsByBillNumber := bills.GetSimilaritySectionsByBillNumber(latestBillItem, sampleSize)

		// This is the equivalent of es_similar_bills_dict in BillMap
		similarBillsDict := bills.GetSimilarBillsDict(similaritySectionsByBillNumber, max_bills)
		log.Info().Msgf("Similar Bills Dict: %v", similarBillsDict)
		log.Info().Msgf("Similar Bills Dict Len: %d", len(similarBillsDict))

		// This is a different data form that uses the section metadata as keys
		//similarBillMapBySection := bills.SimilarSectionsItemsToBillMap(similaritySectionsByBillNumber)
		//bills := bills.GetSimilarBills(similarBillMapBySection)
		//log.Info().Msgf("Similar Bills: %v", bills)
		//TODO Select top bills based on score
		//Find how many sections and how many matches
		similarBillsList := make([]string, len(similarBillsDict))

		// TODO this creates the bill list from similars;
		// make sure the original billnumberversion is in the list
		i := 0
		for _, v := range similarBillsDict {
			if len(v) > 0 {
				similarBillsList[i] = v[0].BillCongressTypeNumberVersion
			}
			i++
		}
		similarBillsList = bills.RemoveDuplicates(append(similarBillsList, latestBillItem.BillNumber+latestBillItem.BillVersion))
		compareMatrix := bills.CompareBills(parentPath, similarBillsList, false)
		log.Info().Msgf("Compare Matrix: %v", compareMatrix)
	}
}
