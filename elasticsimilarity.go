package bills

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/elastic/go-elasticsearch/v7"
)

var (
	r             map[string]interface{}
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

	// 3. Search for the indexed documents
	//
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
