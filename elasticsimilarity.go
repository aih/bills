package bills

import (
	"strconv"

	"github.com/rs/zerolog/log"
)

const (
	num_results   = 20 // Maximum number of results to return
	min_sim_score = 20 // Minimum similarity to make a match in the section query
)

func GetSimilaritySectionsByBillNumber(billNumber string) (similarSectionsItems SimilarSectionsItems) {
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
	for sectionIndex, sectionItem := range billsections {
		// The billnumber and billnumber version are not stored in the ES results
		// We add them back in to track the section query with its original bill
		sectionItem.BillNumber = billNumber
		sectionItem.BillNumberVersion = billnumberversion
		sectionItem.SectionIndex = strconv.Itoa(sectionIndex)

		// TODO this can be made concurrent:
		// Send the query out, collect the results put them in order by sectionIndex
		similarSectionsItem := SectionItemQuery(sectionItem)
		similarSectionsItems = append(similarSectionsItems, similarSectionsItem)
	}
	log.Debug().Msgf("number of similarSectionsItems: %d\n", len(similarSectionsItems))
	return similarSectionsItems
}

/*
func SimilarSectionsItemsToBillMap(similarSectionsItems SimilarSectionsItems) (SimilarBillMap SimilarBillMap) {
	// Get bill numbers from similarSectionsItems.SimilarBills and similarSectionsItems.SimilarBillNumberVersions
	// For each bill number, create a map[string]SimilarSections, with the structure below
	// Each item in the slice is the best match, in the target bill, for each section of the original bill
	    Similarity by bill (es_similar_bills_dict)
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
}

func GetSimilarityByBillNumber(billNumber string) (SimilarBillMap SimilarBillMap) {
	return SimilarSectionsItemsToBillMap(GetSimilaritySectionsByBillNumber(billNumber))
}
*/
