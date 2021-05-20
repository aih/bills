package main

import (
	//"context"

	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	bh "github.com/timshannon/badgerhold"
)

type Item struct {
	ID       int
	Name     string
	Category string `badgerholdIndex:"Category"`
	Created  time.Time
}

func main() {
	debug := flag.Bool("debug", false, "sets log level to debug")

	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Debug().Msg("Log level set to Debug")
	flag.Parse()

	options := bh.DefaultOptions
	options.Dir = "data"
	options.ValueDir = "data"

	store, err := bh.Open(options)
	if err != nil {
		// handle error
		log.Fatal().Err(err)
	}
	defer store.Close()

	err = store.Insert(1234, Item{
		Name:    "Test Name",
		Created: time.Now(),
	})

	if err != nil {
		// handle error
		log.Fatal().Err(err)
	}

	var result []Item

	query := bh.Where("Name").Eq("Test Name")
	err = store.Find(&result, query)

	if err != nil {
		// handle error
		log.Fatal().Err(err)
	}

	fmt.Println(result)

}
