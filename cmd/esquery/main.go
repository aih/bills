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
	var (
		billList   BillList
		sampleSize int
		parentPath string
		maxBills   int
	)
	shorthand := " (shorthand)"
	flagBillnumbersUsage := "comma-separated list of billnumbers"
	flag.Var(&billList, "billnumbers", flagBillnumbersUsage)
	flag.Var(&billList, "b", flagBillnumbersUsage+shorthand)
	flag.IntVar(&sampleSize, "samplesize", 0, "number of sections to sample in large bill")
	flagPathUsage := "Absolute path to the parent directory for 'congress' and json metadata files"
	flagPathValue := string(bills.ParentPathDefault)
	flag.StringVar(&parentPath, "parentPath", flagPathValue, flagPathUsage)
	flag.StringVar(&parentPath, "p", flagPathValue, flagPathUsage+shorthand)
	flag.IntVar(&maxBills, "maxBills", max_bills, "maximum number of similar bills to return")

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
		similarBillsDict := bills.GetSimilarBillsDict(similaritySectionsByBillNumber, maxBills)
		log.Debug().Msgf("Similar Bills Dict: %v", similarBillsDict)
		log.Info().Msgf("Similar Bills Dict Len: %d", len(similarBillsDict))

		// This is a different data form that uses the section metadata as keys
		//similarBillMapBySection := bills.SimilarSectionsItemsToBillMap(similaritySectionsByBillNumber)
		//bills := bills.GetSimilarBills(similarBillMapBySection)
		//log.Info().Msgf("Similar Bills: %v", bills)
		//TODO Select top bills based on score
		//Find how many sections and how many matches
		similarBillVersionsList := make([]string, len(similarBillsDict))
		similarBillsList := make([]string, len(similarBillsDict))

		i := 0
		for _, v := range similarBillsDict {
			if len(v) > 0 {
				similarBillVersionsList[i] = v[0].BillCongressTypeNumberVersion
				similarBillsList[i] = v[0].Billnumber
			}
			i++
		}
		// Include the original billnumberversion is in the list if it is not in the list of similar bills
		if index, ok := bills.Find(similarBillsList, billnumber); ok {
			similarBillVersionsList = bills.RemoveIndex(similarBillVersionsList, index)
		}
		similarBillVersionsList = bills.PrependSlice(similarBillVersionsList, latestBillItem.BillNumber+latestBillItem.BillVersion)
		log.Info().Msgf("similar bills: %v", similarBillVersionsList)
		compareMatrix, err := bills.CompareBills(parentPath, similarBillVersionsList, false)
		if err != nil {
			log.Error().Msgf("Error comparing bills: '%v'", err)
		} else {
			log.Info().Msgf("Compare Matrix: %v", compareMatrix)
		}
	}
}
