package bills

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

var (
	batchNum      int
	scrollID      string
	searchIndices = []string{"billsections"}
	idQuery       = map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"fields":  []string{"id"},
		"_source": false,
	}
)

func GetKeysFromMap(m map[string]interface{}) (keys []string) {
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func GetTopHit(hits Hits_ES) (topHit Hit_ES) {

	var topScore float32
	var score float32
	topScore = 0
	for _, item := range hits {
		score = item.Score
		if score > topScore {
			topScore = score
			topHit = item
		}

	}
	return topHit
}

func GetHitsES(results SearchResult_ES) (innerHits Hits_ES, err error) {
	return results.Hits.Hits, nil
}

func GetInnerHits(results SearchResult_ES) (innerHits []InnerHits, err error) {
	var hitsES, _ = GetHitsES(results)
	for _, hit := range hitsES {
		innerHits = append(innerHits, hit.InnerHits)
	}
	return innerHits, nil
}

func GetMatchingBills(results SearchResult_ES) (billnumbers []string) {
	hits, _ := GetHitsES(results)
	for _, item := range hits {
		source := item.Source
		billnumber := source.BillNumber
		billnumbers = append(billnumbers, billnumber)
	}
	return billnumbers
}

func GetMatchingBillNumberVersions(results SearchResult_ES) (billnumberversions []string) {
	hits, _ := GetHitsES(results)
	for _, item := range hits {
		source := item.Source
		billnumber := source.BillNumber
		billversion := source.BillVersion
		billnumberversions = append(billnumberversions, billnumber+billversion)
	}
	return billnumberversions
}

// similars is the result of the MLT query
func GetSimilarSections(results SearchResult_ES, queryItem SectionItem) (similarSections SimilarSections, err error) {
	hits, _ := GetHitsES(results)
	innerHits, _ := GetInnerHits(results)
	for index, hit := range hits {
		var topInnerResultSectionHit InnerHit
		// innerHits follows the same index as hits; for each hit
		// in the top level Hits.Hits, there is an InnerHits array of sections.
		// Of these, only the first one (highest score) is relevant.
		innerResultSectionHits := innerHits[index].Sections.Hits.Hits
		if len(innerResultSectionHits) > 0 {
			// The first section matched is the best section (and usu. the only real match in the bill)
			topInnerResultSectionHit = innerResultSectionHits[0]
		}
		billSource := hit.Source
		similarSection := SimilarSection{
			Billnumber:                    billSource.BillNumber,
			BillCongressTypeNumberVersion: billSource.ID,
			Congress:                      billSource.Congress,
			Session:                       billSource.Session,
			Legisnum:                      billSource.Legisnum,
			Score:                         topInnerResultSectionHit.Score,
			SectionIndex:                  topInnerResultSectionHit.Source.SectionIndex,
			SectionNum:                    topInnerResultSectionHit.Source.SectionNumber + " ",
			SectionHeader:                 topInnerResultSectionHit.Source.SectionHeader,
			TargetSectionHeader:           queryItem.SectionHeader,
			TargetSectionNumber:           queryItem.SectionNumber + " ",
			TargetSectionIndex:            queryItem.SectionIndex,
			Date:                          billSource.Date,
		}
		dublinCores := billSource.DC
		dublinCore := ""
		if len(dublinCores) > 0 {
			dublinCore = dublinCores[0]
			result := DcTitle_Regexp.FindAllStringSubmatch(dublinCore, -1)
			title := ""
			if len(result) > 0 && len(result[0]) > 1 {
				title = strings.Trim(result[0][1], " ")
			}
			similarSection.Title = title
			similarSections = append(similarSections, similarSection)
		}
	}
	sort.SliceStable(similarSections, func(i, j int) bool { return similarSections[i].Score > similarSections[j].Score })

	return similarSections, nil
}

func SectionItemQuery(sectionItem SectionItem) (similarSectionsItem SimilarSectionsItem) {
	log.Debug().Msgf("Get similar sections for: '%s'", sectionItem.SectionHeader)
	sectionText := sectionItem.SectionText
	esResult, err := GetMLTResult(num_results, min_sim_score, sectionText)

	if err != nil {
		log.Error().Msgf("Error getting results: '%v'", err)
	}
	//bs, _ := json.Marshal(similars)
	//fmt.Println(string(bs))
	//ioutil.WriteFile("similarsResp.json", bs, os.ModePerm)

	hitsEs, _ := GetHitsES(esResult) // = Hits.Hits
	hitsLen := len(hitsEs)

	log.Debug().Msgf("Number of bills with matching sections (hitsLen): %d\n", hitsLen)
	innerHits, _ := GetInnerHits(esResult) // = InnerHits for each hit of Hits.Hits
	var sectionHitsLen int
	for index, hit := range innerHits {
		billHit := hitsEs[index]
		log.Debug().Msg("\n===============\n")
		log.Debug().Msgf("Bill %d of %d", index+1, hitsLen)
		log.Debug().Msgf("Matching sections for: %s", billHit.Source.BillNumber+billHit.Source.BillVersion)
		log.Debug().Msgf("Score for %s: %f", billHit.Source.BillNumber, billHit.Score)
		log.Debug().Msg("\n******************\n")
		sectionHits := hit.Sections.Hits.Hits
		sectionHitsLen = len(sectionHits)
		log.Debug().Msgf("sectionHitsLen: %d\n", sectionHitsLen)
		for _, sectionHit := range sectionHits {
			log.Debug().Msgf("sectionMatch: %s", sectionHit.Source.SectionHeader)
			log.Debug().Msgf("Section score: %f", sectionHit.Score)
		}
		log.Debug().Msg("\n******************\n")

	}
	//log.Debug().Msgf("similarSections: %v\n", similarSections)
	var matchingBillsDedupe []string
	var matchingBillNumberVersionsDedupe []string
	if len(innerHits) > 0 {
		topHit := GetTopHit(hitsEs)
		matchingBills := GetMatchingBills(esResult)
		matchingBillsDedupe = RemoveDuplicates(matchingBills)
		matchingBillsString := strings.Join(matchingBills, ", ")

		log.Debug().Msgf("Number of matches: %d, MatchingBills: %s, MatchingBillsDedupe: %s, Top Match: %s, Score: %f", len(innerHits), matchingBillsString, matchingBillsDedupe, topHit.Source.BillNumber, topHit.Score)

		matchingBillNumberVersions := GetMatchingBillNumberVersions(esResult)
		matchingBillNumberVersionsDedupe = RemoveDuplicates(matchingBillNumberVersions)
		matchingBillNumberVersionsString := strings.Join(matchingBillNumberVersionsDedupe, ", ")

		log.Debug().Msgf("Number of matching bill versions: %d, Matches: %s", len(matchingBillNumberVersionsDedupe), matchingBillNumberVersionsString)
	}
	similarSections, _ := GetSimilarSections(esResult, sectionItem)
	log.Debug().Msgf("number of similarSections: %v\n", len(similarSections))
	log.Debug().Msgf("sectionIndex: %v\n", sectionItem.SectionIndex)
	return SimilarSectionsItem{
		BillNumber:                sectionItem.BillNumber,
		BillNumberVersion:         sectionItem.BillNumberVersion,
		SectionHeader:             sectionItem.SectionHeader,
		SectionNum:                sectionItem.SectionNumber,
		SectionIndex:              sectionItem.SectionIndex,
		SimilarSections:           similarSections,
		SimilarBills:              matchingBillsDedupe,
		SimilarBillNumberVersions: matchingBillNumberVersionsDedupe,
	}

}

func ReadToString(r io.Reader) string {
	var b bytes.Buffer
	b.ReadFrom(r)
	return b.String()
}

func PrintESInfo() {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatal().Msgf("Error creating the client: %s", err)
	}
	res, err := es.Info()
	if err != nil {
		log.Fatal().Msgf("Error getting response: %s", err)
	} else {
		log.Info().Msg(fmt.Sprint(res))
		log.Info().Msg(fmt.Sprint(elasticsearch.Version))
	}
}

func MakeMLTQuery(size, minscore int, searchtext string) (mltquery map[string]interface{}) {
	mltquery = map[string]interface{}{
		"size":      20,
		"min_score": 15,
		"query": map[string]interface{}{
			"nested": map[string]interface{}{
				"path": "sections",
				"query": map[string]interface{}{
					"more_like_this": map[string]interface{}{
						"fields":          []string{"sections.section_text"},
						"like":            `SEC. 102. COLORADO WILDERNESS ADDITIONS. (a) Designation.—Section 2(a) of the Colorado Wilderness Act of 1993 (16 U.S.C. 1132 note; Public Law 103–77) is amended— (1) in paragraph (18), by striking “1993,” and inserting “1993, and certain Federal land within the White River National Forest that comprises approximately 6,896 acres, as generally depicted as ‘Proposed Ptarmigan Peak Wilderness Additions’ on the map entitled ‘Proposed Ptarmigan Peak Wilderness Additions’ and dated June 24, 2019,”; and (2) by adding at the end the following:`,
						"min_term_freq":   2,
						"max_query_terms": 60,
						"min_doc_freq":    2,
					},
				},
				"inner_hits": map[string]interface{}{
					"highlight": map[string]interface{}{
						"fields": map[string]interface{}{
							"sections.section_text": map[string]interface{}{},
						},
					},
				},
			},
		},
	}

	mltquery["size"] = size
	mltquery["min_score"] = minscore
	mltquery["query"].(map[string]interface{})["nested"].(map[string]interface{})["query"].(map[string]interface{})["more_like_this"].(map[string]interface{})["like"] = searchtext
	return mltquery
}

func MakeBillQuery(billnumber string) (billquery map[string]interface{}) {
	billquery = map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"billnumber": billnumber,
			},
		},
	}
	return billquery
}

func RunQuery(query map[string]interface{}) (r map[string]interface{}) {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatal().Msgf("Error creating the client: %s", err)
	}

	// Search for the indexed documents
	// Build the request body.
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatal().Msgf("Error encoding query: %s", err)
	}

	// Perform the search request.
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(searchIndices...),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatal().Msgf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatal().Msgf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Fatal().Msgf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatal().Msgf("Error parsing the response body: %s", err)
	}

	if res.Status() != "200 OK" {
		log.Error().Msgf("ES search: [%s]", res.Status())
	}

	// Print the response status, number of results, and request duration.
	log.Debug().Msgf(
		"ES search: [%s] %d hits; took %dms",
		res.Status(),
		int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(r["took"].(float64)),
	)
	// Print the ID and document source for each hit.
	/*
		for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
			log.Debug().Msgf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
		}
	*/
	return r
}

func GetBill_ES(billnumber string) map[string]interface{} {
	r := RunQuery(MakeBillQuery(billnumber))
	return r
}

func BillResultToStruct(billresult map[string]interface{}) (billItemResult BillItemES, err error) {
	bs, _ := json.Marshal(billresult)
	if err := json.Unmarshal([]byte(bs), &billItemResult); err != nil {
		log.Error().Msgf("Could not parse ES bill query result: %v", err)
	}
	return billItemResult, err
}

func GetMoreLikeThis_ES(size, minscore int, searchtext string) map[string]interface{} {
	r := RunQuery(MakeMLTQuery(size, minscore, searchtext))
	return r
}

func GetMLTResult(size, minscore int, searchtext string) (esResult SearchResult_ES, err error) {

	similars := GetMoreLikeThis_ES(num_results, min_sim_score, searchtext)

	// TODO: marshal and unmarshal is not efficient, but the mapstructure library does not work for this
	bs, _ := json.Marshal(similars)
	if err := json.Unmarshal([]byte(bs), &esResult); err != nil {
		log.Error().Msgf("Could not parse ES query result: %v", err)
	}
	return esResult, err
}

// Performs scroll query over indices in `searchIndices`; sends result to the resultChan for processing to extract billnumbers
// See https://github.com/elastic/go-elasticsearch/issues/44#issuecomment-483974031
func ScrollQueryBillNumbers(buf bytes.Buffer, resultChan chan []gjson.Result) {
	defer close(resultChan)
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatal().Msgf("Error creating the client: %s", err)
	}

	// Perform the initial search request to get
	// the first batch of data and the scroll ID
	//
	log.Info().Msg("Scrolling the index...")
	log.Info().Msg(strings.Repeat("-", 80))
	res, _ := es.Search(
		es.Search.WithIndex(searchIndices...),
		es.Search.WithBody(&buf),
		es.Search.WithSort("_doc"),
		es.Search.WithSize(10000),
		es.Search.WithScroll(time.Minute),
	)

	// Handle the first batch of data and extract the scrollID
	//
	json := ReadToString(res.Body)
	res.Body.Close()
	//fmt.Println(json)

	scrollID = gjson.Get(json, "_scroll_id").String()

	log.Debug().Msg("Batch   " + strconv.Itoa(batchNum))
	log.Debug().Msg("ScrollID: " + scrollID)
	billNumbers := gjson.Get(json, "hits.hits.#fields.id").Array()
	//log.Debug().Msg("IDs:     " + strings.Join(billNumbers, ", "))
	resultChan <- billNumbers
	log.Debug().Msg(strings.Repeat("-", 80))

	// Perform the scroll requests in sequence
	//
	for {
		log.Info().Msg("Getting the next batch")
		batchNum++

		// Perform the scroll request and pass the scrollID and scroll duration
		//
		res, err := es.Scroll(es.Scroll.WithScrollID(scrollID), es.Scroll.WithScroll(time.Minute))
		if err != nil {
			log.Fatal().Msgf("Error: %s", err)
		}
		if res.IsError() {
			log.Fatal().Msgf("Error response: %s", res)
		}

		json := ReadToString(res.Body)
		res.Body.Close()

		// Extract the scrollID from response
		//
		scrollID = gjson.Get(json, "_scroll_id").String()

		// Extract the search results
		//
		hits := gjson.Get(json, "hits.hits")

		// Break out of the loop when there are no results
		//
		if len(hits.Array()) < 1 {
			log.Info().Msg("Finished scrolling")
			break
		} else {
			log.Debug().Msg("Batch   " + strconv.Itoa(batchNum))
			log.Debug().Msg("ScrollID: " + scrollID)
			billNumbers := gjson.Get(json, "hits.hits.#.fields.id").Array()
			//log.Debug().Msg("IDs:     " + strings.Join(billNumbers, ", "))
			resultChan <- billNumbers
			log.Debug().Msg(strings.Repeat("-", 80))
		}
	}
}

// Sort the eh, es, and enr as latest
// Then sort by date
// TODO: better method is to get the latest version in Fdsys_billstatus
func GetLatestBill(r map[string]interface{}) (latestbill BillItemES, err error) {
	var latestbillInterface map[string]interface{}
	latestdate, _ := time.Parse(time.RFC3339, time.RFC3339)
	latestbillversion := "ih"
	latestbillversion_val := 0
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		billversion := hit.(map[string]interface{})["_source"].(map[string]interface{})["billversion"].(string)
		datestring := hit.(map[string]interface{})["_source"].(map[string]interface{})["date"]
		if datestring == nil {
			datestring = ""
		}
		if datestring != "" {
			date, err := time.Parse(time.RFC3339, datestring.(string)+"T15:04:05Z")
			if err != nil {
				fmt.Println(err)
			}

			// Use the date if the latest version is not an "e" version
			if date.After(latestdate) && !strings.HasPrefix(latestbillversion, "e") {
				latestdate = date
				latestbillversion = billversion
				latestbillversion_val = BillVersionsOrdered[latestbillversion]
				latestbillInterface = hit.(map[string]interface{})
			}
		}
		if billversion_val, ok := BillVersionsOrdered[billversion]; ok {
			if strings.HasPrefix(billversion, "e") && (billversion_val > latestbillversion_val) {
				latestbillversion = billversion
				latestbillInterface = hit.(map[string]interface{})
				latestbillversion_val = BillVersionsOrdered[latestbillversion]
			}
		}
		log.Debug().Msgf("bill=%s; date=%s", billversion, datestring)
		log.Debug().Msgf("current latestbillversion=%s; latestdate=%s, latestbillversionval=%d", latestbillversion, latestdate.String(), latestbillversion_val)
	}
	log.Debug().Msgf("latestbillversion=%s; latestdate=%s, latestbillversionval=%d", latestbillversion, latestdate.String(), latestbillversion_val)
	var billItem BillItemES
	if latestbillInterface["_source"] != nil {
		billItem, err = BillResultToStruct(latestbillInterface["_source"].(map[string]interface{}))
		if err != nil {
			log.Fatal().Msgf("Error converting bill to struct: %s", err)
		}
	} else {
		log.Error().Msg("Bill item is not found")
	}
	return billItem, err
}

func GetSampleBillNumbers() []string {
	// TODO: set number of bills to get and get them at random
	return []string{"116hr299"}
}

// Gets all ids, which includes bill and version
func GetAllBillNumbers() []string {
	var billNumbers []gjson.Result
	resultChan := make(chan []gjson.Result)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(idQuery); err != nil {
		log.Fatal().Msgf("Error encoding query: %s", err)
	}
	ScrollQueryBillNumbers(buf, resultChan)
	for newBillNumbers := range resultChan {
		billNumbers = append(billNumbers, newBillNumbers...)
	}
	//fmt.Println(billNumbers)
	// billNumbers is an Array of gjson.Result;
	// each result is itself an array of string of the form
	//["117hr141ih"]
	log.Info().Msgf("Length of billNumbers: %d", len(billNumbers))
	var billNumberStrings []string
	for _, b := range billNumbers {
		bRes := b.Array()
		for _, bItem := range bRes {
			billNumber := bItem.String()
			if billNumber != "" {
				billNumberStrings = append(billNumberStrings, billNumber)
			}
		}
	}
	return billNumberStrings
}
