package bills

import (
	"testing"

	"github.com/aih/bills/internal/testutils"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

const billPath1 = "/path/to/data/116/bills/hr/hr1500/text-versions/rh/document.xml"

const billPath2 = "/path/to/congress/116/bills/hr222/BILLS-116hr222ih-uslm.xml"

func TestBillPathRegex(t *testing.T) {
	log.Info().Msg("Test getting billnumber_version from bill path")
	testutils.SetLogLevel()
	var billnumber1 = BillNumberFromPath(billPath1)
	assert.Equal(t, "116hr1500rh", billnumber1)

	//if path1 != "116hr1500rh" {
	//	t.Errorf("billnumber not extracted correctly from bill path: %s", billPath1)
	//}
	var billnumber2 = BillNumberFromPath(billPath2)
	assert.Equal(t, "116hr222ih", billnumber2)
}
