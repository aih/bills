package bills

import (
	"testing"

	"github.com/aih/bills/internal/testutils"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestDownloadCommitteesYaml(t *testing.T) {
	log.Info().Msg("Test downloading Committees YAML from unitedstates repo")
	testutils.SetLogLevel()
	_, err := DownloadCommitteesYaml()
	if err != nil {
		t.Errorf("Error downloading Committees YAML: %v", err)
	}

}

func TestReadCommitteesYaml(t *testing.T) {
	log.Info().Msg("Test parsing YAML to committees list")
	testutils.SetLogLevel()
	committees, err := ReadCommitteesYaml()
	if err != nil {
		t.Errorf("Could not read Committees YAML from file: %v", err)
	}
	if committees.Committees == nil || len(committees.Committees) == 0 {
		t.Errorf("No committees found in YAML file")
	}
	assert.GreaterOrEqual(t, len(committees.Committees), 1)
	assert.Equal(t, committees.Committees[0].Type, "house")
}
