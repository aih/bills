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
func GetSimilaritySectionsByBillNumber(billItem BillItemES, samplesize int) (similarSectionsItems SimilarSectionsItems) {
	billversion := billItem.BillVersion
	billNumber := billItem.BillNumber
	billnumberversion := billNumber + billversion
	billsections := billItem.Sections

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

func GetSimilarityBillMapBySection(billItem BillItemES, sampleSize int) (similarBillMapBySection SimilarBillMapBySection) {
	return SimilarSectionsItemsToBillMap(GetSimilaritySectionsByBillNumber(billItem, sampleSize))
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

func GetSimilarBillsDict(similarSectionsItems SimilarSectionsItems, maxBills int) (similarBillsDict map[string]SimilarSections) {
	var sectionSimilars []SimilarSections
	var similarBillsAll []string
	similarBillsDict = make(map[string]SimilarSections)
	dedupeMap := make(map[SectionItemMeta]SimilarSection)
	/*
		SectionItemMeta
		BillNumber        string `json:"bill_number"`
		BillNumberVersion string `json:"bill_number_version"`
		SectionIndex      string `json:"sectionIndex"`
		SectionNumber     string `json:"section_number"`
		SectionHeader
	*/
	for _, similarSectionsItem := range similarSectionsItems {
		// Collect the similar sections
		sectionSimilars = append(sectionSimilars, similarSectionsItem.SimilarSections)
		similarBillsAll = append(similarBillsAll, similarSectionsItem.SimilarBills...)
	}
	similarBillsAll = RemoveDuplicates(similarBillsAll)
	if maxBills > 0 && len(similarBillsAll) > maxBills {
		similarBillsAll = similarBillsAll[:maxBills]
	}
	for _, sectionSimilar := range sectionSimilars {
		for _, similarSection := range sectionSimilar {
			// Check which bill it belongs to
			// Check if that bill already has an item with that TargetSectionNumber
			// If not, add it to the dict
			// If yes, check if the score is higher than the existing one
			// If yes, replace the existing one
			dedupeItemKey := SectionItemMeta{
				BillNumber:        similarSection.Billnumber,
				BillNumberVersion: similarSection.BillCongressTypeNumberVersion,
				SectionNumber:     similarSection.TargetSectionNumber,
				SectionHeader:     similarSection.TargetSectionHeader,
			}
			if _, ok := dedupeMap[dedupeItemKey]; !ok {
				dedupeMap[dedupeItemKey] = similarSection
			} else {
				if similarSection.Score > dedupeMap[dedupeItemKey].Score {
					dedupeMap[dedupeItemKey] = similarSection
				}
			}

		}
	}
	for _, similarSection := range dedupeMap {
		//Check if the bill is in the similarBillsAll list
		if _, ok := Find(similarBillsAll, similarSection.Billnumber); ok {
			similarBillsDict[similarSection.Billnumber] = append(similarBillsDict[similarSection.Billnumber], similarSection)
		}
	}
	return similarBillsDict
}

/*
SimilarBillMap
def getSimilarBills(es_similarity: List[dict] ) -> dict:
similarBills = {}
  sectionSimilars = [item.get('similars', []) for item in es_similarity]
  billnumbers = list(unique_everseen(flatten([[similarItem.get('billnumber') for similarItem in similars] for similars in sectionSimilars])))
  for billnumber in billnumbers:
    try:
      similarBills[billnumber] = []
      for sectionIndex, similarItem in enumerate(sectionSimilars):
        sectionBillItems = sorted(filter(lambda x: x.get('billnumber', '') == billnumber, similarItem), key=lambda k: k.get('score', 0), reverse=True)
        if sectionBillItems and len(sectionBillItems) > 0:
          for sectionBillItem in sectionBillItems:
            # Check if we've seen this billItem before and which has a higher score
            currentScore = sectionBillItem.get('score', 0)
            currentSection = sectionBillItem.get('section_num', '') + sectionBillItem.get('section_header', '')
            dupeIndexes = [similarBillIndex for similarBillIndex, similarBill in enumerate(similarBills.get(billnumber, [])) if (similarBill.get('section_num', '') + similarBill.get('section_header', '')) == currentSection]
            if not dupeIndexes:
              sectionBillItem['sectionIndex'] = str(sectionIndex)
              sectionBillItem['target_section_number'] = es_similarity[sectionIndex].get('section_number', '')
              sectionBillItem['target_section_header'] = es_similarity[sectionIndex].get('section_header', '')
              similarBills[billnumber].append(sectionBillItem)
              break
            elif  currentScore > similarBills[billnumber][dupeIndexes[0]].get('score', 0):
              del similarBills[billnumber][dupeIndexes[0]]
              similarBills[billnumber].append(sectionBillItem)
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
