package bills

import (
	"strings"

	"github.com/rs/zerolog/log"
)

const (
	num_results   = 20 // Maximum number of results to return
	min_sim_score = 20 // Minimum similarity to make a match in the section query
)

func GetSimilaritySectionsByBillNumber(billNumber string) (similarSectionsItems SimilarSectionsItems, err error) {
	log.Info().Msgf("Get versions of: %s", billNumber)
	r := GetBill_ES(billNumber)
	latestBillItem, err := GetLatestBill(r)
	if err != nil {
		log.Error().Msgf("Error getting latest bill: '%v'", err)
	}
	billversion := latestBillItem.BillVersion
	billnumberversion := billNumber + billversion
	billsections := latestBillItem.Sections
	log.Info().Msgf("Get similar bills for the %d sections of bill %s", len(billsections), billnumberversion)
	for _, sectionItem := range billsections {
		log.Info().Msgf("Get similar sections for: '%s'", sectionItem.SectionHeader)
		sectionText := sectionItem.SectionText
		esResult, err := GetMLTResult(num_results, min_sim_score, sectionText)

		if err != nil {
			log.Error().Msgf("Error getting results: '%v'", err)
		}
		//bs, _ := json.Marshal(similars)
		//fmt.Println(string(bs))
		//ioutil.WriteFile("similarsResp.json", bs, os.ModePerm)

		hitsEs, _ := GetHitsES(esResult) // = Hits.Hits
		hitsLen := len(hitsEs)

		log.Info().Msgf("Number of bills with matching sections (hitsLen): %d\n", hitsLen)
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
		similarSectionsItems = append(similarSectionsItems, SimilarSectionsItem{
			BillNumber:        billNumber,
			BillNumberVersion: billnumberversion,
			SectionHeader:     sectionItem.SectionHeader,
			SectionNum:        sectionItem.SectionNumber,
			SimilarSections:   similarSections,
		})
		log.Debug().Msgf("number of similarSections: %v\n", len(similarSections))
		//log.Debug().Msgf("similarSections: %v\n", similarSections)
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
	log.Debug().Msgf("number of similarSectionsItems: %d\n", len(similarSectionsItems))
	return similarSectionsItems, err
}

// Similarity by bill (es_similar_bills_dict)
/*
"116s238": [
	{"date": "2019-01-28",
	"score": 92.7196,
	"title": "116 S238 RS: Special Envoy to Monitor and Combat Anti-Semitism Act of 2019",
	"session": "2",
	"congress": "",
	"legisnum": "S. 238",
	"billnumber": "116s238",
	"section_num": "3. ",
	"sectionIndex": "3",
	"section_header": "Monitoring and Combating anti-Semitism",
	"bill_number_version": "116s238rs",
	"target_section_header": "Monitoring and Combating anti-Semitism",
	"target_section_number": "3."}],
*/
