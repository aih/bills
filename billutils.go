package bills

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
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

//  Gets bill path from the billnumber + version
//  E.g. billnumber of the form 116hr1500rh returns [path]/data/116/bills/hr/hr1/text-versions/rh
func PathFromBillNumber(billNumber string) (string, error) {
	var matchMap = FindNamedMatches(BillnumberRegexCompiled, billNumber)
	fmt.Println(matchMap)
	doctypes := "bills"
	stage, ok := matchMap["stage"]
	if ok && stage[1:] == "amdt" {
		doctypes = "amendments"
	}
	if version, ok := matchMap["version"]; ok {
		return path.Join(matchMap["congress"], doctypes, stage, matchMap["stage"]+matchMap["billnumber"], "text-versions", version), nil
	} else {
		return path.Join(matchMap["congress"], doctypes, stage, matchMap["stage"]+matchMap["billnumber"]), errors.New("no version number in path")
	}
}

// Extracts bill titles from a DataJson struct (based on the form in data.json files)
func getBillTitles(dataJson DataJson) map[string][]string {
	titlesMap := map[string][]string{"titles": make([]string, 0), "titlesWholeBill": make([]string, 0)}
	titles := dataJson.Titles
	if len(titles) > 0 {
		for _, titleItem := range titles {
			titlesMap["titles"] = append(titlesMap["titles"], titleItem.Title)
			if !titleItem.IsForPortion {
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
	log.Info().Msgf("Processing: %s\n", billCongressTypeNumber)
	file, err := os.ReadFile(path)
	defer func() {
		log.Info().Msgf("Finished processing: %s\n", billCongressTypeNumber)
		<-sem
	}()
	if err != nil {
		log.Error().Msgf("Error reading data.json: %s", err)
		return err
	}

	var billMeta BillMeta
	var dat DataJson
	_ = json.Unmarshal([]byte(file), &dat)

	billMeta.Actions = dat.Actions
	billMeta.Number = dat.Number
	billMeta.BillType = dat.BillType
	billMeta.Congress = dat.Congress
	billMeta.BillCongressTypeNumber = billCongressTypeNumber
	billMeta.Committees = dat.Committees
	billMeta.Cosponsors = dat.Cosponsors
	billMeta.History = dat.History
	billMeta.ShortTitle = dat.ShortTitle
	titlesMap := getBillTitles(dat)
	billMeta.Titles = titlesMap["titles"]
	billMeta.TitlesWholeBill = titlesMap["titlesWholeBill"]
	billMeta.RelatedBills = dat.RelatedBills
	billMeta.RelatedBillsByBillnumber = make(map[string]RelatedBillItem)
	for i, billItem := range billMeta.RelatedBills {
		if len(billItem.BillId) > 0 {
			log.Debug().Msgf("BCTN: %s\n", BillIdToBillNumber(billItem.BillId))
			billMeta.RelatedBills[i].BillCongressTypeNumber = BillIdToBillNumber(billItem.BillId)
			billMeta.RelatedBillsByBillnumber[billMeta.RelatedBills[i].BillCongressTypeNumber] = billItem
		} else {
			billMeta.RelatedBills[i].BillCongressTypeNumber = ""
		}
	}
	log.Debug().Msgf("billMeta: %v\n", billMeta)
	billMetaStorageChannel <- billMeta
	return nil
}
