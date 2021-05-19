package main

import (
	//"context"

	"fmt"
	"log"
	"time"

	bh "github.com/timshannon/badgerhold"
)

type Item struct {
	ID       int
	Name     string
	Category string `badgerholdIndex:"Category"`
	Created  time.Time
}

func main() {
	options := bh.DefaultOptions
	options.Dir = "data"
	options.ValueDir = "data"

	store, err := bh.Open(options)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	defer store.Close()

	err = store.Insert(1234, Item{
		Name:    "Test Name",
		Created: time.Now(),
	})

	if err != nil {
		// handle error
		log.Fatal(err)
	}

	var result []Item

	query := bh.Where("Name").Eq("Test Name")
	err = store.Find(&result, query)

	if err != nil {
		// handle error
		log.Fatal(err)
	}

	fmt.Println(result)

}
