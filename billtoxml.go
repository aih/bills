package bills

import (
	"fmt"
	"os"

	"github.com/antchfx/xmlquery"
	"github.com/rs/zerolog/log"
)

type BillLevels struct {
	sections []*xmlquery.Node
	levels   []*xmlquery.Node
}

func ParseBill(sampleFilePath string) (parsedBill BillLevels) {
	xmlFile, err := os.Open(sampleFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	doc, err := xmlquery.Parse(xmlFile)
	if err != nil {
		fmt.Println("Error parsing file:", err)
		return
	}
	defer xmlFile.Close()

	sections := xmlquery.Find(doc, "//section")
	levels := xmlquery.Find(doc, "//level")

	parsedBill = BillLevels{sections, levels}

	if len(levels) > 0 {
		for _, level := range levels {
			log.Debug().Msg(level.OutputXML(true))
		}
	} else {
		log.Debug().Msg("No 'level' elements in this document")
	}

	for _, section := range sections {
		log.Debug().Msg(section.OutputXML(true))
	}
	return parsedBill
}
