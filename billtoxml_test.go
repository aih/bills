package bills

import (
	"path"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var sampleFilePath = path.Join("samples", "BILLS-116hr1500eh.xml")

const section12 = `<section id="HC417BF8D57CF48F3841216EC05FBD460"><enum>12.</enum><header>Maintaining the HMDA Explorer tool and the Public Data Platform API</header><text display-inline="no-display-inline">The Consumer Financial protection Bureau may not retire the HMDA Explorer tool or the Public Data Platform API.</text></section>`

func TestParseBill(t *testing.T) {
	log.Info().Msg("Test bill parsing (to sections and levels)")
	setLogLevel()
	billLevels := ParseBill(sampleFilePath)
	gotnumsections := len(billLevels.sections)
	if gotnumsections != 18 {
		t.Errorf("Got %d sections; want 18", gotnumsections)
	}
	gotsection12 := billLevels.sections[11].OutputXML(true)
	if gotsection12 != section12 {
		t.Errorf("For section 12 got: " + gotsection12)
	}

}

func setLogLevel() {
	// Log level set to info
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
