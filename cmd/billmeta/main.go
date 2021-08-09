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

// Command-line function to process and save metadata, with flags for paths.
// Walks the 'congress' directory of the `parentPath`. Runs the following:
// bills.MakeBillsMeta(parentPath) to create bill metadata and store it in a sync file and JSON files for: bills, titlesJson and billMeta
// bills.LoadTitles(bills.TitleNoYearSyncMap, bills.BillMetaSyncMap) to create an index of bill titles without year info
// loadMainTitles(bills.MainTitleNoYearSyncMap, bills.BillMetaSyncMap) to create an index of main bill titles without year info
// bills.WriteBillMetaFiles writes `billMeta.json` in each bill directory
// and then finally writes the whole meta sync file to a single JSON file, billMetaGo.json

// Creates three metadata files: bills, titlesJson and billMeta
func main() {
	// See https://stackoverflow.com/a/55324723/628748
	// Ensure we exit with an error code and log message
	// when needed after deferred cleanups have run.
	// Credit: https://medium.com/@matryer/golang-advent-calendar-day-three-fatally-exiting-a-command-line-tool-with-grace-874befeb64a4
	var err error
	defer func() {
		if err != nil {
			log.Fatal()
		}
	}()

	flagDefs := map[string]flagDef{
		"billMetaPath": {string(bills.BillMetaPath), "Absolute path to store the bill json metadata file"},
		"parentPath":   {string(bills.ParentPathDefault), "Absolute path to the parent directory for 'congress' and json metadata files"},
		"billNumber":   {"", "Get and print billMeta for one bill"},
		"log":          {"Info", "Sets Log level. Options: Error, Info, Debug"},
	}

	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	var parentPath string
	var pathToBillMeta string
	flag.StringVar(&parentPath, "parentPath", flagDefs["parentPath"].value, flagDefs["parentPath"].usage)
	flag.StringVar(&parentPath, "p", flagDefs["parentPath"].value, flagDefs["parentPath"].usage+" (shorthand)")
	flag.StringVar(&pathToBillMeta, "billMetaPath", flagDefs["billMetaPath"].value, flagDefs["billMetaPath"].usage)
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

	// Processes a single bill, based on the bill number
	if billNumber != "" {
		billPath, err := bills.PathFromBillNumber(billNumber)
		if err != nil {
			log.Error().Msgf("Error getting path from billnumber: %s", billNumber)
			return
		}
		billPath = strings.ReplaceAll(billPath, "/text-versions", "")
		bills.MakeBillMeta(parentPath, billPath)
		return
	}

	bills.MakeBillsMeta(parentPath)
	bills.LoadTitles(bills.TitleNoYearSyncMap, bills.BillMetaSyncMap)
	bills.LoadMainTitles(bills.MainTitleNoYearSyncMap, bills.BillMetaSyncMap)
	billlist := getSyncMapKeys(bills.BillMetaSyncMap)
	log.Info().Msgf("BillMetaSyncMap keys: %v", billlist)
	log.Info().Msgf("BillMetaSyncMap length: %v", len(strings.Split(billlist, ", ")))
	bills.WriteBillMetaFiles(bills.BillMetaSyncMap, parentPath)
	log.Info().Msgf("pathToBillMeta: %v", pathToBillMeta)
	if pathToBillMeta == "" {
		if parentPath != "" {
			pathToBillMeta = path.Join(parentPath, bills.BillMetaFile)
		} else {
			pathToBillMeta = bills.BillMetaPath
		}
	}
	log.Info().Msg("Creating string from  billMetaSyncMap")
	jsonString, err := bills.MarshalJSONBillMeta(bills.BillMetaSyncMap)
	if err != nil {
		log.Error().Msgf("Error making JSON data for billMetaMap: %s", err)
	}
	log.Info().Msgf("Writing billMeta JSON data to file: %v", pathToBillMeta)
	os.WriteFile(pathToBillMeta, []byte(jsonString), 0666)

	log.Info().Msgf("BillMetaSyncMap: %v", bills.BillMetaSyncMap)
	billslist := bills.GetSyncMapKeys(bills.BillMetaSyncMap)
	billsString, err := json.Marshal(billslist)
	log.Info().Msgf("Bills: %s", billsString)
	if err != nil {
		log.Error().Msgf("Error making JSON data for bills: %s", err)
	}
	log.Info().Msg("Writing bills JSON data to file")
	currentBillsPath := path.Join(parentPath, bills.BillsFile)
	os.WriteFile(currentBillsPath, []byte(billsString), 0666)

	jsonSimString, err := bills.MarshalJSONBillSimilarity(bills.BillMetaSyncMap)
	if err != nil {
		log.Error().Msgf("Error making JSON data for billSimilarity file: %s", err)
	}
	log.Info().Msg("Writing billSimilarity JSON data to file")
	currentBillSimilarityPath := path.Join(parentPath, bills.BillSimilarityFile)
	os.WriteFile(currentBillSimilarityPath, []byte(jsonSimString), 0666)

	jsonTitleNoYearString, err := bills.MarshalJSONStringArray(bills.TitleNoYearSyncMap)
	if err != nil {
		log.Error().Msgf("Error making JSON data for TitleNoYearSyncMap: %s", err)
	}
	log.Info().Msg("Writing titleNoYearIndex JSON data to file")
	currentTitleNoYearIndexPath := path.Join(parentPath, bills.TitleNoYearIndex)
	os.WriteFile(currentTitleNoYearIndexPath, []byte(jsonTitleNoYearString), 0666)
	jsonMainTitleNoYearString, err := bills.MarshalJSONStringArray(bills.MainTitleNoYearSyncMap)

	if err != nil {
		log.Error().Msgf("Error making JSON data for MainTitleNoYearMap: %s", err)
	}
	log.Info().Msg("Writing maintitleNoYearIndex JSON data to file")
	currentMainTitleNoYearIndexPath := path.Join(parentPath, bills.MainTitleNoYearIndex)
	os.WriteFile(currentMainTitleNoYearIndexPath, []byte(jsonMainTitleNoYearString), 0666)
}
