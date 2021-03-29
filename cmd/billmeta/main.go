package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

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

// Walks the 'congress' directory
// Creates three metadata files: bills, titlesJson and billMeta
// bills is the list of bill numbers (billCongressTypeNumber)
// titles is a list of titles (no year)
// billMeta collects metadata from data.json files
func makeBillMeta() {
	defer fmt.Println("Done")
	billMetaStorageChannel := make(chan bills.BillMeta)
	fmt.Printf("Getting all files in %s.  This may take a while.\n", bills.PathToCongressDataDir)
	dataJsonFiles, _ := bills.ListDataJsonFiles()
	wg := &sync.WaitGroup{}
	wg.Add(len(dataJsonFiles))
	go func() {
		wg.Wait()
		close(billMetaStorageChannel)
	}()

	for _, path := range dataJsonFiles {
		go bills.ExtractBillMeta(path, billMetaStorageChannel, wg)
	}

	billCounter := 0
	for billMeta := range billMetaStorageChannel {
		billCounter++
		fmt.Printf("[%d] Storing metadata for %s.\n", billCounter, billMeta.BillCongressTypeNumber)
		bills.BillMetaSyncMap.Store(billMeta.BillCongressTypeNumber, billMeta)
		for _, title := range billMeta.Titles {
			// titleSyncMap.Store(title, billMeta.BillCongressTypeNumber)
			titleNoYear := bills.TitleNoYearRegexCompiled.ReplaceAllString(title, "")
			if titleBills, loaded := bills.TitleNoYearSyncMap.LoadOrStore(titleNoYear, []string{billMeta.BillCongressTypeNumber}); loaded {
				titleBills = bills.RemoveDuplicates(append(titleBills.([]string), billMeta.BillCongressTypeNumber))
				bills.TitleNoYearSyncMap.Store(titleNoYear, titleBills)
			}

		}
	}
	billslist := getSyncMapKeys(bills.BillMetaSyncMap)
	billsString, err := json.Marshal(billslist)
	if err != nil {
		fmt.Printf("Error making JSON data for bills: %s", err)
	}
	fmt.Println("Writing bills JSON data to file")
	os.WriteFile(bills.BillsPath, []byte(billsString), 0666)

	// Loop through titles and for each bill update relatedbills:
	//  * If the related bill does not already exist, create it
	//  * If the related bill already exists, add the title to the titles array
	//  * Update the "reason" to add "title match"

	//titles := getKeys(titleNoYearSyncMap)
	bills.TitleNoYearSyncMap.Range(func(billTitle, titleBills interface{}) bool {
		for _, titleBill := range titleBills.([]string) {
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
						relatedBillItem.Reason = strings.Join(bills.RemoveDuplicates(append(strings.Split(relatedBillItem.Reason, ", "), bills.TitleMatchReason)), ", ")
						relatedBillItem.Titles = bills.RemoveDuplicates(append(relatedBillItem.Titles, titleBillRelated))
					} else {
						newRelatedBillItem := new(bills.RelatedBillItem)
						newRelatedBillItem.BillCongressTypeNumber = titleBill
						newRelatedBillItem.Titles = []string{billTitle.(string)}
						newRelatedBillItem.Reason = bills.TitleMatchReason
						relatedBills[titleBill] = *newRelatedBillItem
						// TODO add sponsor and cosponsor information to newRelatedBillItem
					}
				}
				// TODO Store new relatedbills
			} else {
				fmt.Printf("No metadata in BillMetaSyncMap for bill: %s", titleBill)
			}

		}
		return true
	})
	jsonString, err := MarshalJSONBillMeta(bills.BillMetaSyncMap)
	if err != nil {
		fmt.Printf("Error making JSON data for billMetaMap: %s", err)
	}
	fmt.Println("Writing billMeta JSON data to file")
	os.WriteFile(bills.BillMetaPath, []byte(jsonString), 0666)
	jsonTitleNoYearString, err := MarshalJSONStringArray(bills.TitleNoYearSyncMap)
	if err != nil {
		fmt.Printf("Error making JSON data for billMetaMap: %s", err)
	}
	fmt.Println("Writing titleNoYearIndex JSON data to file")
	os.WriteFile(bills.TitleNoYearIndexPath, []byte(jsonTitleNoYearString), 0666)
}

func main() {
	makeBillMeta()
}
