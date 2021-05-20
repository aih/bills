package bills

import (
	"log"

	"github.com/elastic/go-elasticsearch/v7"
)

var (
	getAllQuery = `{"query": {"match_all" : {}},"size": 2}`
)

func printESInfo() {
	log.SetFlags(0)
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	} else {
		log.Println(res)
		log.Println(elasticsearch.Version)
	}
}

/*
func getAllDocs() {
	log.SetFlags(0)
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	// Set up the request object.
	req := esapi.IndexRequest{
		Index:      "billsections",
		DocumentID: "115hconres76ih",
		Body:       strings.NewReader(getAllQuery),
		Refresh:    "true",
	}

	// Perform the request with the client.
	_, err = req.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}

	docs := 0
	for {
		res, err := scroller.Do(context.TODO())
		if err == io.EOF {
			// No remaining documents matching the search so break out of the 'forever' loop
			break
		}
		for _, hit := range res.Hits.Hits {
			// JSON parse or do whatever with each document retrieved from your index
			item := make(map[string]interface{})
			err := json.Unmarshal(*hit.Source, &item)
			if err != nil {
				log.Err().Err(err)
			}
			docs++
		}
	}

}
*/
