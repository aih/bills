package bills

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"

	"github.com/elastic/go-elasticsearch/v7"
)

var (
	r             map[string]interface{}
	batchNum      int
	scrollID      string
	searchIndices = []string{"billsections"}
)

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

func SampleQuery() {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatal().Msgf("Error creating the client: %s", err)
	}

	// Search for the indexed documents
	// Build the request body.
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"billnumber": "115hr4134",
			},
		},
	}
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
	// Print the response status, number of results, and request duration.
	fmt.Printf(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(r["took"].(float64)),
	)
	// Print the ID and document source for each hit.
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		fmt.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
	}

}

func getIdQuery() bytes.Buffer {

	// Search indexed documents with a `match_all` query to retrieve all
	// Build the request body.
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"fields":  []string{"id"},
		"_source": false,
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatal().Msgf("Error encoding query: %s", err)
	}
	return buf

}

// Performs scroll query over indices in `searchIndices`; sends result to the resultChan for processing
func scrollQuery(buf bytes.Buffer, resultChan chan []gjson.Result) {
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
	//buf := queryFunc(es)
	res, _ := es.Search(
		es.Search.WithIndex(searchIndices...),
		es.Search.WithBody(&buf),
		es.Search.WithSort("_doc"),
		es.Search.WithSize(10000),
		es.Search.WithScroll(time.Minute),
	)

	// Handle the first batch of data and extract the scrollID
	//
	json := read(res.Body)
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

		json := read(res.Body)
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

func GetAllBillNumbers() []string {
	var billNumbers []gjson.Result
	resultChan := make(chan []gjson.Result)
	buf := getIdQuery()
	go scrollQuery(buf, resultChan)
	for newBillNumbers := range resultChan {
		billNumbers = append(billNumbers, newBillNumbers...)
	}
	//fmt.Println(billNumbers)
	// billNumbers is an Array of gjson.Result;
	// each result is itself an array of string of the formo
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

func read(r io.Reader) string {
	var b bytes.Buffer
	b.ReadFrom(r)
	return b.String()
}
