package bills

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

// Converts a bill_id of the form `hr299-116` into `116hr299`
func BillIdToBillNumber(billId string) string {
	billIdParts := strings.Split(billId, "-")
	return fmt.Sprintf("%s%s", billIdParts[1], billIdParts[0])
}

//  Gets billnumber + version from the bill path
//  E.g. bill_path of the form e.g. [path]/data/116/bills/hr/hr1/text-versions
//    returns 116hr1500rh
func BillNumberFromPath(billPath string) string {
	var matchMap = FindNamedMatches(UsCongressPathRegexCompiled, billPath)
	if version, ok := matchMap["version"]; ok {
		return fmt.Sprintf("%s%s%s", matchMap["congress"], matchMap["billnumber"], version)
	} else {
		return fmt.Sprintf("%s%s", matchMap["congress"], matchMap["billnumber"])
	}
}

// Extracts bill titles from a DataJson struct (based on the form in data.json files)
func getBillTitles(dataJson DataJson) map[string][]string {
	titlesMap := map[string][]string{"titles": make([]string, 0), "titlesWholeBill": make([]string, 0)}
	titles := dataJson.Titles
	if len(titles) > 0 {
		for _, titleItem := range titles {
			titlesMap["titles"] = append(titlesMap["titles"], titleItem.Title)
			if titleItem.IsForPortion {
				titlesMap["titlesWholeBill"] = append(titlesMap["titlesWholeBill"], titleItem.Title)
			}
		}
	}
	return titlesMap

}

// Extracts bill metadata from a path to a data.json file; sends it to the billMetaStorageChannel
// as part of a WaitGroup passed as wg
func ExtractBillMeta(path string, billMetaStorageChannel chan BillMeta, sem chan bool, wg *sync.WaitGroup) error {
	defer wg.Done()

	billCongressTypeNumber := BillNumberFromPath(path)
	fmt.Printf("Processing: %s\n", billCongressTypeNumber)
	file, err := os.ReadFile(path)
	<-sem
	if err != nil {
		log.Printf("Error reading data.json: %s", err)
		return err
	}

	var billMeta BillMeta
	var dat DataJson
	_ = json.Unmarshal([]byte(file), &dat)

	billMeta.BillCongressTypeNumber = billCongressTypeNumber
	billMeta.Committees = dat.Committees
	billMeta.Cosponsors = dat.Cosponsors
	titlesMap := getBillTitles(dat)
	billMeta.Titles = titlesMap["titles"]
	billMeta.TitlesWholeBill = titlesMap["titlesWholeBill"]
	billMeta.RelatedBills = dat.RelatedBills
	billMeta.RelatedBillsByBillnumber = make(map[string]RelatedBillItem)
	for i, billItem := range billMeta.RelatedBills {
		if len(billItem.BillId) > 0 {
			// fmt.Printf("BCTN: %s\n", BillIdToBillNumber(billItem.BillId))
			billMeta.RelatedBills[i].BillCongressTypeNumber = BillIdToBillNumber(billItem.BillId)
			billMeta.RelatedBillsByBillnumber[billMeta.RelatedBills[i].BillCongressTypeNumber] = billItem
		} else {
			billMeta.RelatedBills[i].BillCongressTypeNumber = ""
		}
	}
	//fmt.Printf("billMeta: %v\n", billMeta)
	billMetaStorageChannel <- billMeta
	return nil
}
