package bills

import (
	"strings"
	"testing"

	"github.com/aih/bills/internal/testutils"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestMakeBillMeta(t *testing.T) {
	testutils.SetLogLevel()
	log.Info().Msg("Test getting metadata from bill by bill number")
	billNumber := "116hr1500"
	billPath, err := PathFromBillNumber(billNumber)
	if err != nil {
		log.Error().Msgf("Error getting path from billnumber: %s", billNumber)
		return
	}
	billPath = strings.ReplaceAll(billPath, "/text-versions", "")
	billMeta := MakeBillMeta("./samples", billPath)
	// Tests that we get a correct ShortTitle field from the metadata
	assert.Equal(t, "Consumers First Act", billMeta.ShortTitle)
	assert.Equal(t, 2, len(billMeta.TitlesWholeBill))
	assert.Equal(t, 29, len(billMeta.Cosponsors))
}

func TestMakeBillsMeta(t *testing.T) {
	testutils.SetLogLevel()
	log.Info().Msg("Test gathering metadata from a directory")
	parentPath := "./samples"
	MakeBillsMeta(parentPath)

	// Tests that we get a correct ShortTitle field from the metadata
	assert.FileExists(t, parentPath+"/billsGo.json")
}

//TODO make tests for MakeBillsMeta, LoadTitles, LoadMainTitles
