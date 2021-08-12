package bills

import (
	"encoding/json"
	"strings"

	"github.com/rs/zerolog/log"
)

const (
	num_results   = 20
	min_sim_score = 20
)

func GetSimilarityByBillNumber(billNumber string) (esResults []SearchResult_ES) {
	log.Info().Msgf("Get versions of: %s", billNumber)
	r := GetBill_ES(billNumber)
	latestbill := GetLatestBill(r)
	billversion := latestbill["_source"].(map[string]interface{})["billversion"].(string)
	billnumberversion := billNumber + billversion
	billsections := latestbill["_source"].(map[string]interface{})["sections"].([]interface{})
	log.Info().Msgf("Get similar bills for the %d sections of bill %s", len(billsections), billnumberversion)
	for _, sectionItem := range billsections {
		sectionItemInterface := sectionItem.(map[string]interface{})
		log.Info().Msgf("Get similar sections for: '%s'", sectionItemInterface["section_header"])
		sectionText := sectionItemInterface["section_text"]
		similars := GetMoreLikeThisQuery(num_results, min_sim_score, sectionText.(string))

		// TODO: marshal and unmarshal is not efficient, but the mapstructure library does not work for this
		var esResult SearchResult_ES
		bs, _ := json.Marshal(similars)
		if err := json.Unmarshal([]byte(bs), &esResult); err != nil {
			log.Error().Msgf("Could not parse ES query result: %v", err)
		} else {
			esResults = append(esResults, esResult)
		}
		//bs, _ := json.Marshal(similars)
		//fmt.Println(string(bs))
		//ioutil.WriteFile("similarsResp.json", bs, os.ModePerm)

		hitsEs, _ := GetHitsES(esResult) // = Hits.Hits
		hitsLen := len(hitsEs)

		// TODO: Organize the matches by bill and by section
		// Define a struct for the similarity JSON per bill

		log.Debug().Msgf("hitsLen: %d\n", hitsLen)
		innerHits, _ := GetInnerHits(esResult) // = InnerHits for each hit of Hits.Hits
		var sectionHitsLen int
		for index, hit := range innerHits {
			billHit := hitsEs[index]
			log.Debug().Msg("\n===============\n")
			log.Debug().Msgf("Bill %d of %d", index+1, hitsLen)
			log.Debug().Msgf("Matching sections for: %s", billHit.Source.BillNumber+billHit.Source.BillVersion)
			log.Debug().Msgf("Score for %s: %f", billHit.Source.BillNumber, billHit.Score)
			log.Debug().Msg("\n******************\n")
			sectionHits := hit.Sections.Hits.Hits
			sectionHitsLen = len(sectionHits)
			log.Debug().Msgf("sectionHitsLen: %d\n", sectionHitsLen)
			for _, sectionHit := range sectionHits {
				log.Debug().Msgf("sectionMatch: %s", sectionHit.Source.SectionHeader)
				log.Debug().Msgf("Section score: %f", sectionHit.Score)
			}
			log.Debug().Msg("\n******************\n")

		}
		similarSections, _ := GetSimilarSections(esResult)
		log.Debug().Msgf("number of similarSections: %v\n", len(similarSections))
		log.Debug().Msgf("similarSections: %v\n", similarSections)
		if len(innerHits) > 0 {
			topHit := GetTopHit(hitsEs)
			matchingBills := GetMatchingBills(esResult)
			matchingBillsDedupe := RemoveDuplicates(matchingBills)
			matchingBillsString := strings.Join(matchingBills, ", ")

			log.Debug().Msgf("Number of matches: %d, Matches: %s, MatchesDedupe: %s, Top Match: %s, Score: %f", len(innerHits), matchingBillsString, matchingBillsDedupe, topHit.Source.BillNumber, topHit.Score)

			matchingBillNumberVersions := GetMatchingBillNumberVersions(esResult)
			matchingBillNumberVersionsDedupe := RemoveDuplicates(matchingBillNumberVersions)
			matchingBillNumberVersionsString := strings.Join(matchingBillNumberVersionsDedupe, ", ")

			log.Debug().Msgf("Number of matches: %d, Matches: %s", len(innerHits), matchingBillNumberVersionsString)
		}
	}
	return esResults
}
