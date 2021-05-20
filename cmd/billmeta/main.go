package main

import (
	"encoding/json"
	"flag"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/aih/bills"
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

// Marshals a sync.Map object of the type map[string]BillMeta
// see https://stackoverflow.com/a/46390611/628748
// and https://stackoverflow.com/a/65442862/628748
func MarshalJSONBillMeta(m *sync.Map) ([]byte, error) {
	tmpMap := make(map[string]bills.BillMeta)
	m.Range(func(k interface{}, v interface{}) bool {
		tmpMap[k.(string)] = v.(bills.BillMeta)
		return true
	})
	return json.Marshal(tmpMap)
}

func MarshalJSONBillSimilarity(m *sync.Map) ([]byte, error) {

	tmpMap := make(map[string][]bills.RelatedBillItem)
	m.Range(func(k interface{}, v interface{}) bool {
		tmpMap[k.(string)] = v.(bills.BillMeta).RelatedBills
		return true
	})
	return json.Marshal(tmpMap)
}

// Marshals a sync.Map object of the type map[string]string
// see https://stackoverflow.com/a/46390611/628748
// and https://stackoverflow.com/a/65442862/628748
func MarshalJSONStringArray(m *sync.Map) ([]byte, error) {
	tmpMap := make(map[string][]string)
	m.Range(func(k interface{}, v interface{}) bool {
		tmpMap[k.(string)] = v.([]string)
		return true
	})
	return json.Marshal(tmpMap)
}

func reverse(ss []string) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}

// Walks the 'congress' directory
// Creates three metadata files: bills, titlesJson and billMeta
// bills is the list of bill numbers (billCongressTypeNumber)
// titles is a list of titles (no year)
// billMeta collects metadata from data.json files
func makeBillMeta(parentPath string) {
	pathToBillMeta := bills.BillMetaPath
	pathToCongressDir := bills.PathToCongressDataDir
	if parentPath != "" {
		pathToBillMeta = path.Join(parentPath, bills.BillMetaFile)
		pathToCongressDir = path.Join(parentPath, bills.CongressDir)
	}
	defer log.Info().Msg("Done")
	// Limiting openfiles prevents memory issues
	// See http://jmoiron.net/blog/limiting-concurrency-in-go/
	maxopenfiles := 100
	sem := make(chan bool, maxopenfiles)
	billMetaStorageChannel := make(chan bills.BillMeta)
	log.Info().Msgf("Getting all files in %s.  This may take a while.", pathToCongressDir)
	dataJsonFiles, _ := bills.ListDataJsonFiles(pathToCongressDir)
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
			billCounter++
			log.Info().Msgf("[%d] Storing metadata for %s.", billCounter, billMeta.BillCongressTypeNumber)
			// Get related bill data
			bills.BillMetaSyncMap.Store(billMeta.BillCongressTypeNumber, billMeta)
			saveErr := bills.SaveBillJson(billMeta.BillCongressTypeNumber, billMeta)
			if saveErr != nil {
				log.Print(saveErr)
			}
			/* Saves each bill JSON to an item in db
			saveDbErr := bills.SaveBillJsonToDB(billMeta.BillCongressTypeNumber, billMeta)
			if saveDbErr != nil {
				log.Info().Msg(saveDbErr)
			}
			*/

			// Add 	billMeta.ShortTitle to billMeta.Titles
			shortTitle := billMeta.ShortTitle
			titles := billMeta.Titles
			if shortTitle != "" {
				titles = append(billMeta.Titles, billMeta.ShortTitle)
			}
			for _, title := range titles {
				//for _, title := range billMeta.Titles {
				log.Info().Msgf("[%d] Getting titles for %s.", billCounter, billMeta.BillCongressTypeNumber)
				titleNoYear := bills.TitleNoYearRegexCompiled.ReplaceAllString(title, "")
				if titleBills, loaded := bills.TitleNoYearSyncMap.LoadOrStore(titleNoYear, []string{billMeta.BillCongressTypeNumber}); loaded {
					titleBills = bills.RemoveDuplicates(append(titleBills.([]string), billMeta.BillCongressTypeNumber))
					bills.TitleNoYearSyncMap.Store(titleNoYear, titleBills)
				}

			}
		}
	}()

	for _, path := range dataJsonFiles {
		sem <- true
		go bills.ExtractBillMeta(path, billMetaStorageChannel, sem, wg)
	}

	billslist := getSyncMapKeys(bills.BillMetaSyncMap)
	billsString, err := json.Marshal(billslist)
	if err != nil {
		log.Error().Msgf("Error making JSON data for bills: %s", err)
	}
	log.Info().Msg("Writing bills JSON data to file")
	os.WriteFile(bills.BillsPath, []byte(billsString), 0666)

	// Loop through titles and for each bill update relatedbills:
	//  * If the related bill does not already exist, create it
	//  * If the related bill already exists, add the title to the titles array
	//  * Update the "reason" to add "title match"

	log.Info().Msg("***** Processing title matches ******")
	//titles := getKeys(titleNoYearSyncMap)
	bills.TitleNoYearSyncMap.Range(func(billTitle, titleBills interface{}) bool {
		//log.Info().Msg(titleBills)
		for _, titleBill := range titleBills.([]string) {
			//log.Info().Msg("titleBill ", titleBill)
			if billItem, ok := bills.BillMetaSyncMap.Load(titleBill); ok {
				relatedBills := billItem.(bills.BillMeta).RelatedBillsByBillnumber
				if relatedBills != nil && len(relatedBills) > 0 {
					relatedBills = bills.RelatedBillMap{}
				}
				// TODO check that each of titleBills is in relatedBills
				// If it is, make sure 'title match' is one of the reasons
				// Add the billTitle to Titles, if it is not already there
				// If it's not, add it with 'title match'
				for _, titleBillRelated := range titleBills.([]string) {
					if relatedBillItem, ok := relatedBills[titleBillRelated]; ok {
						log.Print("Bill with Related Title ", titleBillRelated)
						relatedBillItem.Reason = strings.Join(bills.SortReasons(bills.RemoveDuplicates(append(strings.Split(relatedBillItem.Reason, ", "), bills.TitleMatchReason))), ", ")
						relatedBillItem.Titles = bills.RemoveDuplicates(append(relatedBillItem.Titles, titleBillRelated))
					} else {
						newRelatedBillItem := new(bills.RelatedBillItem)
						newRelatedBillItem.BillCongressTypeNumber = titleBillRelated
						newRelatedBillItem.Titles = []string{billTitle.(string)}
						newRelatedBillItem.Reason = bills.TitleMatchReason
						relatedBills[titleBillRelated] = *newRelatedBillItem
						// TODO add sponsor and cosponsor information to newRelatedBillItem
					}
				}
				// TODO Store new relatedbills
			} else {
				log.Info().Msgf("No metadata in BillMetaSyncMap for bill: %s", titleBill)
			}

		}
		return true
	})
	log.Info().Msg("Creating string from  billMetaSyncMap")
	jsonString, err := MarshalJSONBillMeta(bills.BillMetaSyncMap)
	if err != nil {
		log.Error().Msgf("Error making JSON data for billMetaMap: %s", err)
	}
	log.Info().Msg("Writing billMeta JSON data to file")
	os.WriteFile(pathToBillMeta, []byte(jsonString), 0666)

	jsonSimString, err := MarshalJSONBillSimilarity(bills.BillMetaSyncMap)
	if err != nil {
		log.Error().Msgf("Error making JSON data for billSimilarity file: %s", err)
	}
	log.Info().Msg("Writing billSimilarity JSON data to file")
	os.WriteFile(bills.BillSimilarityPath, []byte(jsonSimString), 0666)

	jsonTitleNoYearString, err := MarshalJSONStringArray(bills.TitleNoYearSyncMap)
	if err != nil {
		log.Error().Msgf("Error making JSON data for billMetaMap: %s", err)
	}
	log.Info().Msg("Writing titleNoYearIndex JSON data to file")
	os.WriteFile(bills.TitleNoYearIndexPath, []byte(jsonTitleNoYearString), 0666)
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
}

func main() {
	debug := flag.Bool("debug", false, "sets log level to debug")

	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	flagUsage := "Absolute path to the parent directory for 'congress' and json metadata files"
	flagValue := string(bills.ParentPathDefault)
	var parentPath string
	flag.StringVar(&parentPath, "parentPath", flagValue, flagUsage)
	flag.StringVar(&parentPath, "p", flagValue, flagUsage+" (shorthand)")
	flag.Parse()

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	log.Debug().Msg("Log level set to Debug")
	makeBillMeta(parentPath)
}
