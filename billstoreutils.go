package bills

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// Saves bill metadata to billMeta.json
func SaveBillJson(billCongressTypeNumber string, billMetaItem BillMeta) error {

	dataPath, _ := PathFromBillNumber(billCongressTypeNumber)
	dataPath = path.Join(PathToCongressDataDir, "data", strings.Replace(dataPath, "/text-versions", "", 1))
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		return fmt.Errorf("error getting path for: %s\nErr: %s", billCongressTypeNumber, err)
	}
	fmt.Printf("Saving metadata to file for: %s; %s\n", billCongressTypeNumber, dataPath)
	defer func() {
		fmt.Printf("Finished saving metadata for: %s\n", billCongressTypeNumber)
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
