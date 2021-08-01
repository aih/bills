package bills

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

// Converts a bill number of the form `116hr299` into `hr299-116`
func BillNumberToBillId(billNumber string) string {
	log.Debug().Msgf("Billnumber: %s\n", billNumber)
	var matchMap = FindNamedMatches(BillnumberRegexCompiled, billNumber)
	log.Debug().Msg(fmt.Sprint(matchMap))
	return fmt.Sprintf("%s-%s", matchMap["stage"]+matchMap["billnumber"], matchMap["congress"])
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
//  E.g. billnumber of the form 116hr1500rh returns [path]/116/bills/hr/hr1/text-versions/rh
func PathFromBillNumber(billNumber string) (string, error) {
	var matchMap = FindNamedMatches(BillnumberRegexCompiled, billNumber)
	log.Debug().Msg(fmt.Sprint(matchMap))
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

// Unmarshals from JSON to a syncMap
// See https://stackoverflow.com/a/65442862/628748
func UnmarshalJson(data []byte) (*sync.Map, error) {
	var tmpMap map[interface{}]interface{}
	m := &sync.Map{}

	if err := json.Unmarshal(data, &tmpMap); err != nil {
		return m, err
	}

	for key, value := range tmpMap {
		m.Store(key, value)
	}
	return m, nil
}

func UnmarshalJsonFile(jpath string) (*sync.Map, error) {
	jsonFile, err := os.Open(jpath)
	if err != nil {
		log.Error().Err(err)
	}

	defer jsonFile.Close()

	jsonByte, _ := ioutil.ReadAll(jsonFile)
	return UnmarshalJson(jsonByte)
}

// Marshals a sync.Map object of the type map[string]BillMeta
// see https://stackoverflow.com/a/46390611/628748
// and https://stackoverflow.com/a/65442862/628748
func MarshalJSONBillMeta(m *sync.Map) ([]byte, error) {
	tmpMap := make(map[string]BillMeta)
	m.Range(func(k interface{}, v interface{}) bool {
		tmpMap[k.(string)] = v.(BillMeta)
		return true
	})
	return json.Marshal(tmpMap)
}

func MarshalJSONBillSimilarity(m *sync.Map) ([]byte, error) {

	tmpMap := make(map[string][]RelatedBillItem)
	m.Range(func(k interface{}, v interface{}) bool {
		tmpMap[k.(string)] = v.(BillMeta).RelatedBills
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

// Extracts bill metadata from a path to a data.json file; sends it to the billMetaStorageChannel
// as part of a WaitGroup passed as wg
func ExtractBillMeta(billPath string, billMetaStorageChannel chan BillMeta, sem chan bool, wg *sync.WaitGroup) error {
	defer wg.Done()

	billCongressTypeNumber := BillNumberFromPath(billPath)
	log.Info().Msgf("Processing: %s\n", billCongressTypeNumber)
	file, err := os.ReadFile(billPath)
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
	if billMeta.Congress == "" {
		msg := fmt.Sprintf("wrong data in data.json (e.g. no Congress field) for %s", billCongressTypeNumber)
		log.Error().Msg(msg)
		err = errors.New(msg)
		return err
	}
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
