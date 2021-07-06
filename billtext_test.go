package bills

import (
	"fmt"
	"testing"

	"github.com/aih/bills/internal/testutils"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestCollectWordSamplesFromBills(t *testing.T) {
	testutils.SetLogLevel()
	log.Info().Msg("Test collecting word samples from bills")
	wordSamples := CollectWordSamplesFromBills("./samples")
	fmt.Printf("%#v\n", wordSamples)
	// Tests that we get at least 10 words from the sample bills
	assert.Greater(t, len(wordSamples), 10)
}
