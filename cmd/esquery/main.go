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

	billNumbers := bills.GetAllBillNumbers()
	billNumbers = []string{"116hr1500"}
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
			similars := bills.GetMoreLikeThisQuery(5, 20, sectionText.(string))
			hits := similars["hits"].(map[string]interface{})["hits"].([]interface{})
			if len(hits) > 0 {
				log.Info().Msgf("Number of matches: %d, Score (first match): %f", len(hits), hits[0].(map[string]interface{})["_score"])
			}
		}
	}
	// fmt.Print(billNumbers)
}
