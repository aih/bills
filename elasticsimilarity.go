package bills

import (
	"sort"
	"strconv"

	"github.com/rs/zerolog/log"
)

const (
	num_results   = 20 // Maximum number of results to return
	min_sim_score = 25 // Minimum similarity to make a match in the section query
)

func getRandomSliceSectionItems(slice []SectionItem, num_items int) []SectionItem {
	if num_items > len(slice) {
		num_items = len(slice)
	}
	var random_slice []SectionItem
	for i := 0; i < num_items; i++ {
		random_slice = append(random_slice, slice[i])
	}
	return random_slice
}

// Set sample size to <= 0 to use all sections
func GetSimilaritySectionsByBillNumber(billNumber string, samplesize int) (similarSectionsItems SimilarSectionsItems) {
	log.Info().Msgf("Get versions of: %s", billNumber)
	r := GetBill_ES(billNumber)
	log.Info().Msgf("Number of versions of: %s, %d", billNumber, len(r["hits"].(map[string]interface{})["hits"].([]interface{})))
	latestBillItem, err := GetLatestBill(r)
	if err != nil {
		log.Error().Msgf("Error getting latest bill: '%v'", err)
	}
	billversion := latestBillItem.BillVersion
	billnumberversion := billNumber + billversion
	billsections := latestBillItem.Sections
	if samplesize > 0 && len(billsections) > samplesize {
		log.Info().Msgf("Get similar bills for %d of the %d sections of bill %s", samplesize, len(billsections), billnumberversion)
		billsections = getRandomSliceSectionItems(billsections, samplesize)
	} else {
		log.Info().Msgf("Get similar bills for the %d sections of bill %s", len(billsections), billnumberversion)
	}
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

func SimilarSectionsItemsToBillMap(similarSectionsItems SimilarSectionsItems) (similarBillMapBySection SimilarBillMapBySection) {
	// Get bill numbers from similarSectionsItems.SimilarBills and similarSectionsItems.SimilarBillNumberVersions
	// Creates the similarBillMapBySection
	//	  "116s238": {
	//		  SectionItemMeta: SimilarSection
	//	  }
	similarBillMapBySection = make(SimilarBillMapBySection)
	for _, similarSectionsItem := range similarSectionsItems {
		for _, similarSection := range similarSectionsItem.SimilarSections {
			// the inner key is a SectionItemMeta, which can be created from the similarSectionsItem
			innerKey := SectionItemMeta{
				BillNumber:        similarSection.Billnumber,
				BillNumberVersion: similarSection.BillCongressTypeNumberVersion,
				SectionIndex:      similarSectionsItem.SectionIndex,
				SectionNumber:     similarSectionsItem.SectionNum,
				SectionHeader:     similarSectionsItem.SectionHeader,
			}
			if _, ok := similarBillMapBySection[similarSection.Billnumber]; !ok {
				similarBillMapBySection[similarSection.Billnumber] = SimilarBillData{
					SectionItemMetaMap: map[SectionItemMeta]SimilarSection{innerKey: similarSection},
				}
			} else {
				if _, ok := similarBillMapBySection[similarSection.Billnumber].SectionItemMetaMap[innerKey]; !ok {
					similarBillMapBySection[similarSection.Billnumber].SectionItemMetaMap[innerKey] = similarSection
				} else if similarSection.Score > similarBillMapBySection[similarSection.Billnumber].SectionItemMetaMap[innerKey].Score {
					similarBillMapBySection[similarSection.Billnumber].SectionItemMetaMap[innerKey] = similarSection
				}
			}

		}
	}
	log.Info().Msgf("number of items in similarBillMapBySection: %d\n", len(similarBillMapBySection))
	for bill, billdata := range similarBillMapBySection {
		var totalScore float32
		var topSectionScore float32
		var topSectionNum string
		var topSectionHeader string
		var topSectionIndex string
		for _, similarSection := range billdata.SectionItemMetaMap {
			totalScore += similarSection.Score
			if similarSection.Score > topSectionScore {
				topSectionScore = similarSection.Score
				topSectionNum = similarSection.SectionNum
				topSectionHeader = similarSection.SectionNum
				topSectionIndex = similarSection.SectionIndex
			}

		}
		similarBillMapBySection[bill] = SimilarBillData{
			SectionItemMetaMap:   billdata.SectionItemMetaMap,
			TotalScore:           totalScore,
			TopSectionScore:      topSectionScore,
			TopSectionNum:        topSectionNum,
			TopSectionHeader:     topSectionHeader,
			TopSectionIndex:      topSectionIndex,
			TotalSimilarSections: len(billdata.SectionItemMetaMap),
		}

	}
	log.Info().Msgf("similarBillMapBySection: %v\n", similarBillMapBySection)
	return similarBillMapBySection
}

func GetSimilarityBillMapBySection(billNumber string, sampleSize int) (similarBillMapBySection SimilarBillMapBySection) {
	return SimilarSectionsItemsToBillMap(GetSimilaritySectionsByBillNumber(billNumber, sampleSize))
}

type BillScore struct {
	BillNumber      string
	Score           float32
	SectionsMatched int
	// TODO add fields for number of sections matched
}

func GetSimilarBills(similarBillMapBySection SimilarBillMapBySection) (billScores []BillScore) {
	for bill, billdata := range similarBillMapBySection {
		billScores = append(billScores, BillScore{bill, billdata.TotalScore, billdata.TotalSimilarSections})
	}
	sort.SliceStable(billScores, func(i, j int) bool { return billScores[i].Score > billScores[j].Score })
	return billScores
}

/*
SimilarBillMap
For each bill number, create a map[string]SimilarSections, with the structure below
Each item in the slice is the best match, in the target bill, for each section of the original bill
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
		   	"target_section_number": "3."}],...]
*/
