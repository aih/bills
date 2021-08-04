package bills

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/rs/zerolog/log"
	bh "github.com/timshannon/badgerhold"
)

// Saves bill metadata to billMeta.json
func SaveBillJson(billCongressTypeNumber string, billMetaItem BillMeta, parentPath string) error {

	dataPath, _ := PathFromBillNumber(billCongressTypeNumber)

	dataPath = path.Join(parentPath, CongressDir, "data", strings.Replace(dataPath, "/text-versions", "", 1))
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		return fmt.Errorf("error getting path for: %s\nErr: %s", billCongressTypeNumber, err)
	}
	log.Info().Msgf("Saving metadata to file for: %s; %s\n", billCongressTypeNumber, dataPath)
	defer func() {
		log.Info().Msgf("Finished saving metadata for: %s\n", billCongressTypeNumber)
	}()
	file, marshalErr := json.MarshalIndent(billMetaItem, "", " ")
	if marshalErr != nil {
		return fmt.Errorf("error marshalling metadata for: %s\nErr: %s", billCongressTypeNumber, marshalErr)
	}

	writeErr := ioutil.WriteFile(path.Join(dataPath, "billMeta.json"), file, 0644)

	if writeErr != nil {
		return fmt.Errorf("error writing metadata for: %s\nErr: %s", billCongressTypeNumber, writeErr)
	}

	return nil
}

// Saves bill metadata to db (badger or bolt) via bh
func SaveBillJsonToDB(billCongressTypeNumber string, billMetaItem BillMeta) error {
	var options = bh.DefaultOptions
	options.Dir = "data"
	options.ValueDir = "data"

	store, err := bh.Open(options)
	if err != nil {
		// handle error
		log.Fatal().Err(err)
	}
	defer store.Close()

	return store.Insert(billCongressTypeNumber, billMetaItem)
}
