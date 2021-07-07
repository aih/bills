package main

import (
	"flag"
	"os"
	"strings"

	"github.com/aih/bills"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	num_results   = 20
	min_sim_score = 20
)

func getMatchingBills(hits []interface{}) (billnumbers []string) {

	for _, item := range hits {
		source := item.(map[string]interface{})["_source"].(map[string]interface{})
		billnumber := source["billnumber"].(string)
		billnumbers = append(billnumbers, billnumber)
	}
	return billnumbers
}

func main() {
	debug := flag.Bool("debug", false, "sets log level to debug")
	all := flag.Bool("all", false, "processes all bills-- otherwise process a sample")

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
	} else {
		billNumbers = bills.GetSampleBillNumbers()
	}
	for _, billnumber := range billNumbers {
		log.Info().Msgf("Get versions of: %s", billnumber)
		r := bills.GetBill_ES(billnumber)
		latestbill := bills.GetLatestBill(r)
		billversion := latestbill["_source"].(map[string]interface{})["billversion"].(string)
		billnumberversion := billnumber + billversion
		billsections := latestbill["_source"].(map[string]interface{})["sections"].([]interface{})
		log.Info().Msgf("Get similar bills for the %d sections of %s", len(billsections), billnumberversion)
		for _, sectionItem := range billsections {
			//sectionHeader := sectionItem.(map[string]interface{})["section_header"]
			//sectionNumber := sectionItem.(map[string]interface{})["section_number"]
			sectionText := sectionItem.(map[string]interface{})["section_text"]
			similars := bills.GetMoreLikeThisQuery(num_results, min_sim_score, sectionText.(string))
			hits := similars["hits"].(map[string]interface{})["hits"].([]interface{})
			if len(hits) > 0 {
				topHit := bills.GetTopHit(hits)
				matchingBills := strings.Join(getMatchingBills(hits), ", ")

				log.Info().Msgf("Number of matches: %d, Matches: %s, Top Match: %s, Score: %f", len(hits), matchingBills, topHit["_source"].(map[string]interface{})["billnumber"], topHit["_score"])
			}
		}
	}
	// fmt.Print(billNumbers)
}
