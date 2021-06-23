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

type flagDef struct {
	value string
	usage string
}

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

func writeBillMetaFiles(billMetaSyncMap *sync.Map) {
	billMetaSyncMap.Range(func(billCongressTypeNumber, billMeta interface{}) bool {
		saveErr := bills.SaveBillJson(billCongressTypeNumber.(string), billMeta.(bills.BillMeta))
		if saveErr != nil {
			log.Print(saveErr)
		}
		return true
	})
}

func loadTitles(titleSyncMap *sync.Map, billMetaSyncMap *sync.Map) {
	log.Info().Msg("***** Processing title matches ******")

	titleSyncMap.Range(func(billTitle, titleBills interface{}) bool {
		//log.Info().Msg(titleBills)
		for _, titleBill := range titleBills.([]string) {
			//log.Info().Msg("titleBill ", titleBill)
			// titleBill is a bill number
			if billItem, ok := billMetaSyncMap.Load(titleBill); ok {
				billItemStruct := billItem.(bills.BillMeta)
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
						relatedBillItem.Reason = strings.Join(bills.SortReasons(bills.RemoveDuplicates(append(strings.Split(relatedBillItem.Reason, ", "), bills.TitleMatchReason))), ", ")
						relatedBillItem.IdentifiedBy = strings.Join(bills.RemoveDuplicates(append(strings.Split(relatedBillItem.IdentifiedBy, ", "), bills.IdentifiedByBillMap)), ", ")
						relatedBillItem.Titles = bills.RemoveDuplicates(append(relatedBillItem.Titles, billTitle.(string)))
						relatedBills[titleBillRelated] = relatedBillItem
					} else {
						newRelatedBillItem := new(bills.RelatedBillItem)
						newRelatedBillItem.BillCongressTypeNumber = titleBillRelated
						newRelatedBillItem.Titles = []string{billTitle.(string)}
						newRelatedBillItem.Reason = bills.TitleMatchReason
						relatedBillItem.IdentifiedBy = bills.IdentifiedByBillMap
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

func reverse(ss []string) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}

func makeBillMeta(parentPath, billDirPath string) {

	defer log.Info().Msg("Done")
	maxopenfiles := 100
	sem := make(chan bool, maxopenfiles)
	//pathToBillMeta := bills.BillMetaPath
	pathToCongressDir := bills.PathToCongressDataDir
	if parentPath != "" {
		//pathToBillMeta = path.Join(parentPath, bills.BillMetaFile)
		pathToCongressDir = path.Join(parentPath, bills.CongressDir)
	}
	billMetaStorageChannel := make(chan bills.BillMeta)
	billDirFullPath := path.Join(pathToCongressDir, "data", billDirPath)
	log.Info().Msgf("Getting Json for bill: %s", billDirFullPath)
	dataJsonFiles, _ := bills.ListDataJsonFiles(billDirFullPath)
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
		go bills.ExtractBillMeta(jpath, billMetaStorageChannel, sem, wg)
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
func makeBillsMeta(parentPath string) {
	//pathToBillMeta := bills.BillMetaPath
	pathToCongressDir := bills.PathToCongressDataDir
	if parentPath != "" {
		//pathToBillMeta = path.Join(parentPath, bills.BillMetaFile)
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
			if billMeta.Congress == "" {
				continue
			}
			billCounter++
			log.Info().Msgf("[%d] Storing metadata for %s.", billCounter, billMeta.BillCongressTypeNumber)
			// Get related bill data
			bills.BillMetaSyncMap.Store(billMeta.BillCongressTypeNumber, billMeta)
			/* Saves each bill JSON to an item in db
			saveDbErr := bills.SaveBillJsonToDB(billMeta.BillCongressTypeNumber, billMeta)
			if saveDbErr != nil {
				log.Info().Msg(saveDbErr)
			}
			*/

			var mainTitles []string
			// The bill may have one or more of: OfficialTitle, PopularTitle, ShortTitle
			officialTitle := billMeta.OfficialTitle
			shortTitle := billMeta.ShortTitle
			popularTitle := billMeta.PopularTitle
			titles := billMeta.Titles

			if officialTitle != "" {
				mainTitles = append(mainTitles, officialTitle)
			}

			if shortTitle != "" {
				mainTitles = append(mainTitles, shortTitle)
				// Add 	billMeta.ShortTitle to billMeta.Titles
				titles = append(billMeta.Titles, shortTitle)
			}

			if popularTitle != "" {
				mainTitles = append(mainTitles, popularTitle)
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

			for _, title := range mainTitles {
				log.Info().Msgf("[%d] Getting main titles for %s.", billCounter, billMeta.BillCongressTypeNumber)
				mainTitleNoYear := bills.TitleNoYearRegexCompiled.ReplaceAllString(title, "")
				if mainTitleBills, loaded := bills.MainTitleNoYearSyncMap.LoadOrStore(mainTitleNoYear, []string{billMeta.BillCongressTypeNumber}); loaded {
					mainTitleBills = bills.RemoveDuplicates(append(mainTitleBills.([]string), billMeta.BillCongressTypeNumber))
					bills.MainTitleNoYearSyncMap.Store(mainTitleNoYear, mainTitleBills)
				}

			}
		}
	}()

	for _, jpath := range dataJsonFiles {
		sem <- true
		go bills.ExtractBillMeta(jpath, billMetaStorageChannel, sem, wg)
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

	//titles := getKeys(titleNoYearSyncMap)

	/*
		log.Info().Msg("Creating string from  billMetaSyncMap")
		jsonString, err := bills.MarshalJSONBillMeta(bills.BillMetaSyncMap)
		if err != nil {
			log.Error().Msgf("Error making JSON data for billMetaMap: %s", err)
		}
		log.Info().Msg("Writing billMeta JSON data to file")
		os.WriteFile(pathToBillMeta, []byte(jsonString), 0666)
	*/

	jsonSimString, err := bills.MarshalJSONBillSimilarity(bills.BillMetaSyncMap)
	if err != nil {
		log.Error().Msgf("Error making JSON data for billSimilarity file: %s", err)
	}
	log.Info().Msg("Writing billSimilarity JSON data to file")
	os.WriteFile(bills.BillSimilarityPath, []byte(jsonSimString), 0666)

	jsonTitleNoYearString, err := bills.MarshalJSONStringArray(bills.TitleNoYearSyncMap)
	if err != nil {
		log.Error().Msgf("Error making JSON data for TitleNoYearSyncMap: %s", err)
	}
	log.Info().Msg("Writing titleNoYearIndex JSON data to file")
	os.WriteFile(bills.TitleNoYearIndexPath, []byte(jsonTitleNoYearString), 0666)
	jsonMainTitleNoYearString, err := bills.MarshalJSONStringArray(bills.MainTitleNoYearSyncMap)

	if err != nil {
		log.Error().Msgf("Error making JSON data for MainTitleNoYearMap: %s", err)
	}
	log.Info().Msg("Writing maintitleNoYearIndex JSON data to file")
	os.WriteFile(bills.MainTitleNoYearIndexPath, []byte(jsonMainTitleNoYearString), 0666)
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
}

func main() {

	flagDefs := map[string]flagDef{
		"parentPath": {string(bills.ParentPathDefault), "Absolute path to the parent directory for 'congress' and json metadata files"},
		"billNumber": {"", "Get and print billMeta for one bill"},
		"log":        {"Info", "Sets Log level. Options: Error, Info, Debug"},
	}

	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	var parentPath string
	flag.StringVar(&parentPath, "parentPath", flagDefs["parentPath"].value, flagDefs["parentPath"].usage)
	flag.StringVar(&parentPath, "p", flagDefs["parentPath"].value, flagDefs["parentPath"].usage+" (shorthand)")
	debug := flag.Bool("debug", false, "sets log level to debug")

	var logLevel string
	flag.StringVar(&logLevel, "logLevel", flagDefs["log"].value, flagDefs["log"].usage)
	flag.StringVar(&logLevel, "l", flagDefs["log"].value, flagDefs["log"].usage+" (shorthand)")

	var billNumber string
	flag.StringVar(&billNumber, "billNumber", flagDefs["billNumber"].value, flagDefs["billNumber"].usage)

	flag.Parse()

	zLogLevel := bills.ZLogLevels[logLevel]
	log.Info().Msgf("Log Level entered: %s", logLevel)
	log.Info().Msgf("Log Level: %d", zLogLevel)
	// Default is Info (1); Any invalid string will generate Debug (0) level
	zerolog.SetGlobalLevel(zLogLevel)

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	log.Debug().Msg("Log level set to Debug")
	if billNumber != "" {
		billPath, err := bills.PathFromBillNumber(billNumber)
		if err != nil {
			log.Error().Msgf("Error getting path from billnumber: %s", billNumber)
			return
		}
		billPath = strings.ReplaceAll(billPath, "/text-versions", "")
		makeBillMeta(parentPath, billPath)
		return
	}

	makeBillsMeta(parentPath)
	loadTitles(bills.TitleNoYearSyncMap, bills.BillMetaSyncMap)
	writeBillMetaFiles(bills.BillMetaSyncMap)
	pathToBillMeta := bills.BillMetaPath
	if parentPath != "" {
		pathToBillMeta = path.Join(parentPath, bills.BillMetaFile)
	}
	log.Info().Msg("Creating string from  billMetaSyncMap")
	jsonString, err := bills.MarshalJSONBillMeta(bills.BillMetaSyncMap)
	if err != nil {
		log.Error().Msgf("Error making JSON data for billMetaMap: %s", err)
	}
	log.Info().Msg("Writing billMeta JSON data to file")
	os.WriteFile(pathToBillMeta, []byte(jsonString), 0666)
}
