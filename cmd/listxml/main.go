package main

import (
	"os"

	"github.com/aih/bills"
)

func main() {
	pathToCongressDataDir := bills.PathToCongressDataDir
	if len(os.Args) > 1 {
		pathToCongressDataDir = os.Args[1]
	}
	bills.ListDocumentXMLFiles(pathToCongressDataDir)
}
