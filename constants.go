package bills

import (
	"path"
	"regexp"
	"sync"

	"github.com/aih/bills/internal/projectpath"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

type LogLevel int8

type LogLevels map[string]zerolog.Level

// Constants for this package
var (
	BillnumberRegexCompiled = regexp.MustCompile(`(?P<congress>[1-9][0-9]*)(?P<stage>[a-z]{1,8})(?P<billnumber>[1-9][0-9]*)(?P<version>[a-z]+)?`)
	// e.g. congress/data/117/bills/sconres/sconres2
	UsCongressPathRegexCompiled = regexp.MustCompile(`data\/(?P<congress>[1-9][0-9]*)\/(?P<doctype>[a-z]+)\/(?P<stage>[a-z]{1,8})\/(?P<billnumber>[a-z]{1,8}[1-9][0-9]*)\/?(text-versions\/?P<version>[a-z]+)?`)
	// matches strings of the form '...of 1979', where the year is a 4-digit number
	TitleNoYearRegexCompiled = regexp.MustCompile(`of\s[0-9]{4}$`)
	removeXMLRegexCompiled   = regexp.MustCompile(`<[^>]+>`)
	// Set to ../../congress
	PathToDataDir         = path.Join("/", "data")
	ParentPathDefault     = path.Join("..", "..", "..")
	CongressDir           = "congress"
	BillMetaFile          = "billMetaGo.json"
	BillSimilarityFile    = "billSimilarityGo.json"
	TitleNoYearIndex      = "titleNoYearIndexGo.json"
	BillsFile             = "billsGo.json"
	PathToCongressDataDir = path.Join(ParentPathDefault, CongressDir)
	BillMetaPath          = path.Join(ParentPathDefault, BillMetaFile)
	BillSimilarityPath    = path.Join(ParentPathDefault, BillSimilarityFile)
	TitleNoYearIndexPath  = path.Join(ParentPathDefault, TitleNoYearIndex)
	BillsPath             = path.Join(ParentPathDefault, BillsFile)
	BillMetaSyncMap       = new(sync.Map)
	// titleSyncMap                = new(sync.Map)
	TitleNoYearSyncMap  = new(sync.Map)
	TitleMatchReason    = "bills-title_match"
	IdentifiedByBillMap = "BillMap"
	ZLogLevels          = LogLevels{"Debug": zerolog.DebugLevel, "Info": zerolog.InfoLevel, "Error": zerolog.ErrorLevel}
)

func LoadEnv() (err error) {
	return godotenv.Load(path.Join(projectpath.Root, ".env"))
}
