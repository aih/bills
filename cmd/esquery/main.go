package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aih/bills"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	num_results   = 20
	min_sim_score = 20
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

	// allow user to pass billnumbers as argument
	var billList BillList
	flag.Var(&billList, "billnumbers", "comma-separated list of billnumbers")
	flag.Var(&billList, "b", "comma-separated list of billnumbers")
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
		log.Info().Msgf("Get versions of: %s", billnumber)
		r := bills.GetBill_ES(billnumber)
		latestbill := bills.GetLatestBill(r)
		billversion := latestbill["_source"].(map[string]interface{})["billversion"].(string)
		billnumberversion := billnumber + billversion
		billsections := latestbill["_source"].(map[string]interface{})["sections"].([]interface{})
		log.Info().Msgf("Get similar bills for the %d sections of %s", len(billsections), billnumberversion)
		//var lastHit map[string]interface{}
		for _, sectionItem := range billsections {
			//sectionHeader := sectionItem.(map[string]interface{})["section_header"]
			//sectionNumber := sectionItem.(map[string]interface{})["section_number"]
			sectionText := sectionItem.(map[string]interface{})["section_text"]
			similars := bills.GetMoreLikeThisQuery(num_results, min_sim_score, sectionText.(string))

			var esResult bills.SearchResult_ES
			log.Info().Msgf("similars['hits']: %v", similars["hits"])
			mapstructure.Decode(similars, &esResult)
			/*
				body, err := ioutil.ReadAll(similarsRes.Body)
				if err != nil {
					log.Error().Msg("error reading es response body")
					return
				}
			*/

			//bs, _ := json.Marshal(similars)
			//fmt.Println(string(bs))
			//ioutil.WriteFile("similarsResp.json", bs, os.ModePerm)

			log.Info().Msgf("searchResult: %v", esResult)
			//similarsJson := bills.ReadToString(similarsRes.Body)
			//similarsG := gjson.Get(similarsJson, "hits.hits").Array()
			//fmt.Println(string(similarsJson))
			//log.Info().Msgf("similarsBody: %v", similarsRes.Body)
			hits, _ := bills.GetInnerHits(similars)
			//innerResults, _ := bills.GetInnerResults(similars)
			if len(hits) > 0 {
				topHit := bills.GetTopHit(hits)
				matchingBills := strings.Join(getMatchingBills(hits), ", ")
				//lastHit = topHit

				log.Debug().Msgf("Number of matches: %d, Matches: %s, Top Match: %s, Score: %f", len(hits), matchingBills, topHit["_source"].(map[string]interface{})["billnumber"], topHit["_score"])
			}
		}
		//log.Info().Msgf("%v", lastHit["inner_hits"])
		//bs, _ := json.Marshal(lastHit["inner_hits"].(map[string]interface{})["sections"].(map[string]interface{}))
		//fmt.Println(string(bs))
		//source := lastHit["inner_hits"].(map[string]interface{})["sections"].(map[string]interface{})["hits"].(map[string]interface{})["hits"].([]interface{})[0] //.(map[string]interface{})["_source"]

		//bs, _ := json.Marshal(source)
		//fmt.Println(string(bs))
	}
}
