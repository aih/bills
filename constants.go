package bills

import (
	"path"
	"regexp"
	"sync"
)

// Constants for this package
var (
	// e.g. congress/data/117/bills/sconres/sconres2
	UsCongressPathRegexCompiled = regexp.MustCompile(`data\/(?P<congress>[1-9][0-9]*)\/(?P<doctype>[a-z]+)\/(?P<stage>[a-z]{1,8})\/(?P<billnumber>[a-z]{1,8}[1-9][0-9]*)\/?(text-versions\/?P<version>[a-z]+)?`)
	// matches strings of the form '...of 1979', where the year is a 4-digit number
	TitleNoYearRegexCompiled = regexp.MustCompile(`of\s[0-9]{4}$`)
	removeXMLRegexCompiled   = regexp.MustCompile(`<[^>]+>`)
	// Set to ../../congress
	PathToDataDir         = path.Join("/", "data")
	PathToCongressDataDir = path.Join("..", "..", "..", "congress")
	BillMetaPath          = path.Join("..", "..", "..", "billMetaGo.json")
	TitleNoYearIndexPath  = path.Join("..", "..", "..", "titleNoYearIndexGo.json")
	BillsPath             = path.Join("..", "..", "..", "billsGo.json")
	BillMetaSyncMap       = new(sync.Map)
	// titleSyncMap                = new(sync.Map)
	TitleNoYearSyncMap = new(sync.Map)
	TitleMatchReason   = "title match"
)
