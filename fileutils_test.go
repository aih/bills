package bills

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aih/bills/internal/testutils"
	"github.com/rs/zerolog/log"
)

const samplesPath = "./samples"
const samplesPathHR1500 = "samples/congress/data/116/bills/hr/hr1500"

var senateFilesList = []string{"samples/congress/data/117/bills/s/s100", "samples/congress/data/117/bills/s/s100/billMeta.json", "samples/congress/data/117/bills/s/s100/data-fromfdsys-lastmod.txt", "samples/congress/data/117/bills/s/s100/data.json", "samples/congress/data/117/bills/s/s100/data.xml", "samples/congress/data/117/bills/s/s100/fdsys_billstatus-lastmod.txt", "samples/congress/data/117/bills/s/s100/fdsys_billstatus.xml", "samples/congress/data/117/bills/s/s101", "samples/congress/data/117/bills/s/s101/billMeta.json", "samples/congress/data/117/bills/s/s101/data-fromfdsys-lastmod.txt", "samples/congress/data/117/bills/s/s101/data.json", "samples/congress/data/117/bills/s/s101/data.xml", "samples/congress/data/117/bills/s/s101/fdsys_billstatus-lastmod.txt", "samples/congress/data/117/bills/s/s101/fdsys_billstatus.xml"}
var documentXMLFilesSample = []string{"samples/congress/data/116/bills/hr/hr1500/text-versions/eh/document.xml", "samples/congress/data/116/bills/hr/hr1500/text-versions/ih/document.xml", "samples/congress/data/116/bills/hr/hr1500/text-versions/rfs/document.xml", "samples/congress/data/116/bills/hr/hr1500/text-versions/rh/document.xml"}
var dataJsonFilesSample = []string{"samples/congress/data/116/bills/hr/hr1500/data.json"}

var senateBillsFilter = func(testPath string) bool {
	matched, err := regexp.MatchString(`/s[0-9]`, testPath)
	if err != nil {
		return false
	}
	return matched
}

func TestWalkDirFilter(t *testing.T) {
	log.Info().Msg("Test Walking Directory with Filter")
	testutils.SetLogLevel()
	filePaths, err := WalkDirFilter(samplesPath, senateBillsFilter)
	if err != nil {
		t.Errorf("Error walking directory: %v", err)
	}
	assert.ElementsMatch(t, senateFilesList, filePaths)

}
func TestListDocumentXMLFiles(t *testing.T) {
	documentXMLFiles, err := ListDocumentXMLFiles(samplesPathHR1500)
	if err != nil {
		t.Errorf("Error getting document.xml files: %v", err)
	}
	assert.ElementsMatch(t, documentXMLFilesSample, documentXMLFiles)
}
func TestDataJsonFiles(t *testing.T) {
	dataJsonFiles, err := ListDataJsonFiles(samplesPathHR1500)
	if err != nil {
		t.Errorf("Error getting data.json files: %v", err)
	}
	assert.ElementsMatch(t, dataJsonFilesSample, dataJsonFiles)
}
