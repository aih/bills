package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/aih/bills"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	max_bills = 15
)

type SimilarityContext struct {
	ParentPath string
	MaxBills   int
	SampleSize int
	Save       bool
}
type flagDef struct {
	value string
	usage string
}

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

func GetSimilarityForBill(billnumber string, context SimilarityContext) {
	// This is the equivalent of es_similarity in BillMap
	log.Info().Msgf("Get versions of: %s", billnumber)
	r := bills.GetBill_ES(billnumber)
	log.Info().Msgf("Number of versions of %s:, %d", billnumber, len(r["hits"].(map[string]interface{})["hits"].([]interface{})))
	latestBillItem, err := bills.GetLatestBill(r)
	if err != nil {
		log.Error().Msgf("Error getting latest bill: '%v'", err)
	}
	similaritySectionsByBillNumber := bills.GetSimilaritySectionsByBillNumber(latestBillItem, context.SampleSize)
	if context.Save {
		file, marshalErr := json.MarshalIndent(similaritySectionsByBillNumber, "", " ")
		if marshalErr != nil {
			log.Error().Msgf("error marshalling similaritySectionsByBillNumber for: %s\nErr: %s", billnumber, marshalErr)
		}
		log.Info().Msgf("Saving similaritySectionsByBillnumber for: %s\n", billnumber)
		bills.SaveBillDataJson(billnumber, file, context.ParentPath, "esSimilarity.json")
	}

	// This is the equivalent of es_similar_bills_dict in BillMap
	similarBillsDict := bills.GetSimilarBillsDict(similaritySectionsByBillNumber, context.MaxBills)
	log.Debug().Msgf("Similar Bills Dict: %v", similarBillsDict)
	log.Info().Msgf("Similar Bills Dict Len: %d", len(similarBillsDict))
	if context.Save {
		file, marshalErr := json.MarshalIndent(similarBillsDict, "", " ")
		if marshalErr != nil {
			log.Error().Msgf("error marshalling similarBillsDict for: %s\nErr: %s", billnumber, marshalErr)
		}
		bills.SaveBillDataJson(billnumber, file, context.ParentPath, "esSimilarBillsDict.json")
	}

	// This is a different data form that uses the section metadata as keys
	//similarBillMapBySection := bills.SimilarSectionsItemsToBillMap(similaritySectionsByBillNumber)
	//bills := bills.GetSimilarBills(similarBillMapBySection)
	//log.Info().Msgf("Similar Bills: %v", bills)
	//TODO Select top bills based on score
	//Find how many sections and how many matches
	similarBillVersionsList := make([]string, len(similarBillsDict))
	similarBillsList := make([]string, len(similarBillsDict))

	i := 0
	for _, v := range similarBillsDict {
		if len(v) > 0 {
			similarBillVersionsList[i] = v[0].BillCongressTypeNumberVersion
			similarBillsList[i] = v[0].Billnumber
		}
		i++
	}
	// Include the original billnumberversion is in the list if it is not in the list of similar bills
	if index, ok := bills.Find(similarBillsList, billnumber); ok {
		similarBillVersionsList = bills.RemoveIndex(similarBillVersionsList, index)
	}
	similarBillVersionsList = bills.PrependSlice(similarBillVersionsList, latestBillItem.BillNumber+latestBillItem.BillVersion)
	log.Info().Msgf("similar bills: %v", similarBillVersionsList)
	dataPath := path.Join(context.ParentPath, bills.CongressDir, "data")
	compareMatrix, err := bills.CompareBills(dataPath, similarBillVersionsList, false)

	if err != nil {
		log.Error().Msgf("Error comparing bills: '%v'", err)
	} else {
		log.Debug().Msgf("Compare Matrix: %v", compareMatrix)
		// TODO: Save the first row of compare matrix in a file
		compareMap := bills.GetCompareMap(compareMatrix[0])
		compareMapMarshalled, marshalErr := json.MarshalIndent(compareMap, "", " ")
		if marshalErr != nil {
			log.Error().Msgf("error marshalling compareMap for: %s\nErr: %s", billnumber, marshalErr)
		} else {
			bills.SaveBillDataJson(billnumber, compareMapMarshalled, context.ParentPath, "esSimilarCategories.json")
		}
	}
}

func main() {
	save := flag.Bool("save", false, "save results files")
	debug := flag.Bool("debug", false, "sets log level to debug")
	all := flag.Bool("all", false, "processes all bills-- otherwise process a sample")

	// allow user to pass billnumbers as argument
	var (
		billList   BillList
		sampleSize int
		parentPath string
		maxBills   int
		logLevel   string
	)

	shorthand := " (shorthand)"
	flagDefs := map[string]flagDef{
		"billnumbers": {"", "comma-separated list of billnumbers"},
		"parentpath":  {string(bills.ParentPathDefault), "Absolute path to the parent directory for 'congress' and json metadata files"},
		"log":         {"Info", "Sets Log level. Options: Error, Info, Debug"},
	}
	flag.Var(&billList, "b", flagDefs["billnumbers"].usage+shorthand)
	flag.Var(&billList, "billnumbers", flagDefs["billnumbers"].usage)
	flag.IntVar(&sampleSize, "samplesize", 0, "number of sections to sample in large bill")
	flag.StringVar(&parentPath, "parentPath", flagDefs["parentpath"].value, flagDefs["parentpath"].usage)
	flag.StringVar(&parentPath, "p", flagDefs["parentpath"].value, flagDefs["parentpath"].usage+shorthand)
	flag.IntVar(&maxBills, "maxBills", max_bills, "maximum number of similar bills to return")
	flag.StringVar(&logLevel, "logLevel", flagDefs["log"].value, flagDefs["log"].usage)
	flag.StringVar(&logLevel, "l", flagDefs["log"].value, flagDefs["log"].usage+" (shorthand)")

	flag.Parse()
	similarityContext := SimilarityContext{
		ParentPath: parentPath,
		MaxBills:   maxBills,
		SampleSize: sampleSize,
		Save:       *save,
	}

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
	if *save {
		log.Info().Msgf("Saving files to: %s", similarityContext.ParentPath)
	}

	var billNumbers []string
	if *all {
		billNumberVersions := bills.GetAllBillNumbers()
		for _, billNumberVersion := range billNumberVersions {
			billNumber := bills.BillnumberRegexCompiled.ReplaceAllString(billNumberVersion, "$1$2$3")
			billNumbers = append(billNumbers, billNumber)
		}
		billNumbers = bills.RemoveDuplicates(billNumbers)
	} else if len(billList) > 0 {
		billNumbers = billList
	} else {
		billNumbers = bills.GetSampleBillNumbers()
	}

	defer log.Info().Msg("Done")
	/*
		// Limiting openfiles prevents memory issues
		// See http://jmoiron.net/blog/limiting-concurrency-in-go/
		maxopenfiles := 100
		sem := make(chan bool, maxopenfiles)
		// Create three channels: bills.SimilarSectionsItems, bills.SimilarBillsDict, bills.CompareMatrix
		similarSectionsStorageChannel := make(chan bills.SimilarSectionsItems)
		similarBillsDictStorageChannel := make(chan map[string]bills.SimilarSections)
		compareMatrixStorageChannel := make(chan [][]bills.CompareItem)
		wg := &sync.WaitGroup{}
		wg.Add(len(billNumbers))
	*/

	for _, billnumber := range billNumbers {
		GetSimilarityForBill(billnumber, similarityContext)
	}
}
