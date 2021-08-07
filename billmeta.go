package bills

import (
	"encoding/json"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

// Gets keys of a sync.Map
func getSyncMapKeys(m *sync.Map) (s string) {
	m.Range(func(k, v interface{}) bool {
		if s != "" {
			s += ", "
		}
		s += k.(string)
		return true
	})
	return
}

func reverse(ss []string) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}

func LoadTitles(titleSyncMap *sync.Map, billMetaSyncMap *sync.Map) {
	log.Info().Msg("***** Processing title matches ******")
	titleSyncMap.Range(func(billTitle, titleBills interface{}) bool {
		//log.Info().Msg(titleBills)
		for _, titleBill := range titleBills.([]string) {
			//log.Info().Msg("titleBill ", titleBill)
			// titleBill is a bill number
			if billItem, ok := billMetaSyncMap.Load(titleBill); ok {
				billItemStruct := billItem.(BillMeta)
				relatedBills := billItemStruct.RelatedBillsByBillnumber
				/*
					if relatedBills != nil && len(relatedBills) > 0 {
						relatedBills = bills.RelatedBillMap{}
					}
				*/
				// TODO check that each of titleBills is in relatedBills
				// If it is, make sure 'title match' is one of the reasons
				// Add the billTitle to Titles, if it is not already there
				// If it's not, add it with 'title match'
				for _, titleBillRelated := range titleBills.([]string) {
					// titleBillRelated is the bill number of the related bill
					if relatedBillItem, ok := relatedBills[titleBillRelated]; ok {
						log.Debug().Msgf("Bill with Related Title: %s", titleBillRelated)
						relatedBillItem.Reason = strings.Join(SortReasons(RemoveDuplicates(append(strings.Split(relatedBillItem.Reason, ", "), TitleMatchReason))), ", ")
						relatedBillItem.IdentifiedBy = strings.Join(RemoveDuplicates(append(strings.Split(relatedBillItem.IdentifiedBy, ", "), IdentifiedByBillMap)), ", ")
						relatedBillItem.Titles = RemoveDuplicates(append(relatedBillItem.Titles, billTitle.(string)))
						log.Debug().Msgf("Titles: %v", relatedBillItem.Titles)
						relatedBills[titleBillRelated] = relatedBillItem
						if relatedBillItem.BillId == "" && relatedBillItem.BillCongressTypeNumber != "" {
							relatedBillItem.BillId = BillNumberToBillId(relatedBillItem.BillCongressTypeNumber)
						}
						log.Debug().Msgf("relatedBillItem: %v", relatedBillItem)
						if relatedBillItem.BillCongressTypeNumber == "" && relatedBillItem.BillId != "" {
							relatedBillItem.BillCongressTypeNumber = BillIdToBillNumber(relatedBillItem.BillId)
						}
					} else {
						newRelatedBillItem := new(RelatedBillItem)
						newRelatedBillItem.BillCongressTypeNumber = titleBillRelated
						newRelatedBillItem.Titles = []string{billTitle.(string)}
						newRelatedBillItem.Reason = TitleMatchReason
						newRelatedBillItem.IdentifiedBy = IdentifiedByBillMap
						if newRelatedBillItem.BillId == "" && newRelatedBillItem.BillCongressTypeNumber != "" {
							newRelatedBillItem.BillId = BillNumberToBillId(newRelatedBillItem.BillCongressTypeNumber)
						}
						if newRelatedBillItem.BillCongressTypeNumber == "" && newRelatedBillItem.BillId != "" {
							newRelatedBillItem.BillCongressTypeNumber = BillIdToBillNumber(newRelatedBillItem.BillId)
						}
						//if relatedBillItem.Type == "" {
						//}
						relatedBills[titleBillRelated] = *newRelatedBillItem
					}
				}
				// Store new relatedbills
				billItemStruct.RelatedBillsByBillnumber = relatedBills
				billMetaSyncMap.Store(titleBill, billItemStruct)
			} else {
				log.Error().Msgf("No metadata in BillMetaSyncMap for bill: %s", titleBill)
			}

		}
		return true
	})
}

func LoadMainTitles(mainTitleSyncMap *sync.Map, billMetaSyncMap *sync.Map) {
	log.Info().Msg("***** Processing main title matches ******")

	mainTitleSyncMap.Range(func(billTitle, titleBills interface{}) bool {
		//log.Info().Msg(titleBills)
		for _, titleBill := range titleBills.([]string) {
			//log.Info().Msg("titleBill ", titleBill)
			// titleBill is a bill number
			if billItem, ok := billMetaSyncMap.Load(titleBill); ok {
				billItemStruct := billItem.(BillMeta)
				relatedBills := billItemStruct.RelatedBillsByBillnumber
				/*
					if relatedBills != nil && len(relatedBills) > 0 {
						relatedBills = bills.RelatedBillMap{}
					}
				*/
				// Check that each of titleBills is in relatedBills
				// If it is, make sure 'title match' is one of the reasons
				// Add the billTitle to Titles, if it is not already there
				// If it's not, add it with bills.MainTitleMatchReason
				for _, titleBillRelated := range titleBills.([]string) {
					// titleBillRelated is the bill number of the related bill
					if relatedBillItem, ok := relatedBills[titleBillRelated]; ok {
						log.Debug().Msgf("Bill with Related Main Title: %s", titleBillRelated)
						relatedBillItem.Reason = strings.Join(SortReasons(RemoveDuplicates(append(strings.Split(relatedBillItem.Reason, ", "), MainTitleMatchReason))), ", ")
						relatedBillItem.IdentifiedBy = strings.Join(RemoveDuplicates(append(strings.Split(relatedBillItem.IdentifiedBy, ", "), IdentifiedByBillMap)), ", ")
						relatedBillItem.TitlesWholeBill = RemoveDuplicates(append(relatedBillItem.TitlesWholeBill, billTitle.(string)))
						if relatedBillItem.BillId == "" && relatedBillItem.BillCongressTypeNumber != "" {
							relatedBillItem.BillId = BillNumberToBillId(relatedBillItem.BillCongressTypeNumber)
						}
						if relatedBillItem.BillCongressTypeNumber == "" && relatedBillItem.BillId != "" {
							relatedBillItem.BillCongressTypeNumber = BillIdToBillNumber(relatedBillItem.BillId)
						}
						relatedBills[titleBillRelated] = relatedBillItem
					} else {
						newRelatedBillItem := new(RelatedBillItem)
						newRelatedBillItem.BillCongressTypeNumber = titleBillRelated
						newRelatedBillItem.TitlesWholeBill = []string{billTitle.(string)}
						newRelatedBillItem.Reason = MainTitleMatchReason
						relatedBillItem.IdentifiedBy = IdentifiedByBillMap
						if newRelatedBillItem.BillId == "" && newRelatedBillItem.BillCongressTypeNumber != "" {
							newRelatedBillItem.BillId = BillNumberToBillId(newRelatedBillItem.BillCongressTypeNumber)
						}
						if newRelatedBillItem.BillCongressTypeNumber == "" && newRelatedBillItem.BillId != "" {
							newRelatedBillItem.BillCongressTypeNumber = BillIdToBillNumber(newRelatedBillItem.BillId)
						}
						relatedBills[titleBillRelated] = *newRelatedBillItem
					}
				}
				// Store new relatedbills
				billItemStruct.RelatedBillsByBillnumber = relatedBills
				billMetaSyncMap.Store(titleBill, billItemStruct)
			} else {
				log.Error().Msgf("No metadata in BillMetaSyncMap for bill: %s", titleBill)
			}

		}
		return true
	})
}

func WriteBillMetaFiles(billMetaSyncMap *sync.Map, parentPath string) {
	log.Info().Msg("***** Writing individual bill metadata to files ******")

	billMetaSyncMap.Range(func(billCongressTypeNumber, billMeta interface{}) bool {
		log.Info().Msgf("Writing metadata for: %s", billCongressTypeNumber)
		saveErr := SaveBillJson(billCongressTypeNumber.(string), billMeta.(BillMeta), parentPath)
		if saveErr != nil {
			log.Error().Msgf("Error saving meta file: %s", saveErr)
		}
		return true
	})
}

func MakeBillMeta(parentPath, billDirPath string) {

	defer log.Info().Msg("Done")
	maxopenfiles := 100
	sem := make(chan bool, maxopenfiles)
	//pathToBillMeta := bills.BillMetaPath
	pathToCongressDir := PathToCongressDataDir
	if parentPath != "" {
		//pathToBillMeta = path.Join(parentPath, bills.BillMetaFile)
		pathToCongressDir = path.Join(parentPath, CongressDir)
	}
	billMetaStorageChannel := make(chan BillMeta)
	billDirFullPath := path.Join(pathToCongressDir, "data", billDirPath)
	log.Info().Msgf("Getting Json for bill: %s", billDirFullPath)
	dataJsonFiles, _ := ListDataJsonFiles(billDirFullPath)
	wg := &sync.WaitGroup{}
	wg.Add(len(dataJsonFiles))
	go func() {
		wg.Wait()
		close(billMetaStorageChannel)
	}()

	go func() {
		for range dataJsonFiles {
			billMeta := <-billMetaStorageChannel
			log.Debug().Msgf("Got billMeta from Channel: %v\n", billMeta)
		}
	}()

	for _, jpath := range dataJsonFiles {
		sem <- true
		go ExtractBillMeta(jpath, billMetaStorageChannel, sem, wg)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

}

// Walks the 'congress' directory
// Creates three metadata files: bills, titlesJson and billMeta
// bills is the list of bill numbers (billCongressTypeNumber)
// titles is a list of titles (no year)
// billMeta collects metadata from data.json files
func MakeBillsMeta(parentPath string) {
	//pathToBillMeta := BillMetaPath
	pathToCongressDir := PathToCongressDataDir
	if parentPath != "" {
		//pathToBillMeta = path.Join(parentPath, BillMetaFile)
		pathToCongressDir = path.Join(parentPath, CongressDir)
	}
	defer log.Info().Msg("Done")
	// Limiting openfiles prevents memory issues
	// See http://jmoiron.net/blog/limiting-concurrency-in-go/
	maxopenfiles := 100
	sem := make(chan bool, maxopenfiles)
	billMetaStorageChannel := make(chan BillMeta)
	log.Info().Msgf("Getting all files in %s.  This may take a while.", pathToCongressDir)
	dataJsonFiles, _ := ListDataJsonFiles(pathToCongressDir)
	reverse(dataJsonFiles)
	wg := &sync.WaitGroup{}
	wg.Add(len(dataJsonFiles))
	go func() {
		wg.Wait()
		close(billMetaStorageChannel)
	}()
	go func() {
		billCounter := 0
		for range dataJsonFiles {
			billMeta := <-billMetaStorageChannel
			if billMeta.Congress == "" {
				continue
			}
			billCounter++
			log.Info().Msgf("[%d] Storing metadata for %s.", billCounter, billMeta.BillCongressTypeNumber)
			// Get related bill data
			BillMetaSyncMap.Store(billMeta.BillCongressTypeNumber, billMeta)
			/* Saves each bill JSON to an item in db
			saveDbErr := SaveBillJsonToDB(billMeta.BillCongressTypeNumber, billMeta)
			if saveDbErr != nil {
				log.Info().Msg(saveDbErr)
			}
			*/

			var mainTitles []string
			// The bill may have one or more of: OfficialTitle, PopularTitle, ShortTitle
			officialTitle := billMeta.OfficialTitle
			shortTitle := billMeta.ShortTitle
			titles := billMeta.Titles

			if officialTitle != "" {
				mainTitles = RemoveDuplicates(append(mainTitles, officialTitle))
			}

			if shortTitle != "" {
				mainTitles = RemoveDuplicates(append(mainTitles, shortTitle))
				log.Debug().Msgf("Main Titles: %v", mainTitles)
				// Add 	billMeta.ShortTitle to billMeta.Titles
				titles = RemoveDuplicates(append(billMeta.Titles, shortTitle))
				log.Debug().Msgf("Titles: %v", titles)
			}

			for _, title := range titles {
				//for _, title := range billMeta.Titles {
				log.Info().Msgf("[%d] Getting titles for %s.", billCounter, billMeta.BillCongressTypeNumber)
				titleNoYear := strings.Trim(TitleNoYearRegexCompiled.ReplaceAllString(title, ""), " ")
				if titleBills, loaded := TitleNoYearSyncMap.LoadOrStore(titleNoYear, []string{billMeta.BillCongressTypeNumber}); loaded {
					titleBills = RemoveDuplicates(append(titleBills.([]string), billMeta.BillCongressTypeNumber))
					TitleNoYearSyncMap.Store(titleNoYear, titleBills)
				}

			}

			for _, title := range mainTitles {
				log.Info().Msgf("[%d] Getting main titles for %s.", billCounter, billMeta.BillCongressTypeNumber)
				mainTitleNoYear := strings.Trim(TitleNoYearRegexCompiled.ReplaceAllString(title, ""), " ")
				if mainTitleBills, loaded := MainTitleNoYearSyncMap.LoadOrStore(mainTitleNoYear, []string{billMeta.BillCongressTypeNumber}); loaded {
					mainTitleBills = RemoveDuplicates(append(mainTitleBills.([]string), billMeta.BillCongressTypeNumber))
					MainTitleNoYearSyncMap.Store(mainTitleNoYear, mainTitleBills)
				}

			}
		}
	}()

	for _, jpath := range dataJsonFiles {
		sem <- true
		go ExtractBillMeta(jpath, billMetaStorageChannel, sem, wg)
	}

	billslist := getSyncMapKeys(BillMetaSyncMap)
	billsString, err := json.Marshal(billslist)
	if err != nil {
		log.Error().Msgf("Error making JSON data for bills: %s", err)
	}
	log.Info().Msg("Writing bills JSON data to file")
	os.WriteFile(BillsPath, []byte(billsString), 0666)

	// Loop through titles and for each bill update relatedbills:
	//  * If the related bill does not already exist, create it
	//  * If the related bill already exists, add the title to the titles array
	//  * Update the "reason" to add "title match"

	//titles := getKeys(titleNoYearSyncMap)

	/*
		log.Info().Msg("Creating string from  billMetaSyncMap")
		jsonString, err := MarshalJSONBillMeta(BillMetaSyncMap)
		if err != nil {
			log.Error().Msgf("Error making JSON data for billMetaMap: %s", err)
		}
		log.Info().Msg("Writing billMeta JSON data to file")
		os.WriteFile(pathToBillMeta, []byte(jsonString), 0666)
	*/

	jsonSimString, err := MarshalJSONBillSimilarity(BillMetaSyncMap)
	if err != nil {
		log.Error().Msgf("Error making JSON data for billSimilarity file: %s", err)
	}
	log.Info().Msg("Writing billSimilarity JSON data to file")
	os.WriteFile(BillSimilarityPath, []byte(jsonSimString), 0666)

	jsonTitleNoYearString, err := MarshalJSONStringArray(TitleNoYearSyncMap)
	if err != nil {
		log.Error().Msgf("Error making JSON data for TitleNoYearSyncMap: %s", err)
	}
	log.Info().Msg("Writing titleNoYearIndex JSON data to file")
	os.WriteFile(TitleNoYearIndexPath, []byte(jsonTitleNoYearString), 0666)
	jsonMainTitleNoYearString, err := MarshalJSONStringArray(MainTitleNoYearSyncMap)

	if err != nil {
		log.Error().Msgf("Error making JSON data for MainTitleNoYearMap: %s", err)
	}
	log.Info().Msg("Writing maintitleNoYearIndex JSON data to file")
	os.WriteFile(MainTitleNoYearIndexPath, []byte(jsonMainTitleNoYearString), 0666)
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
}
