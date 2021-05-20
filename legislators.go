package bills

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/rs/zerolog/log"
)

var (
	legislatorYamlUrl = "https://raw.githubusercontent.com/unitedstates/congress-legislators/master/legislators-current.yaml"
)

type Legislators struct {
	Legislators []Legislator `yaml:"legislators,omitempty,flow"`
}

type Id struct {
	Id []Legislator `yaml:"id,omitempty,flow"`
}

type Legislator struct {
	Bioguide       string   `yaml:"bioguide,omitempty,flow"`
	Thomas         string   `yaml:"thomas,omitempty,flow"`
	Lis            string   `yaml:"lis,omitempty,flow"`
	Govtrack       string   `yaml:"govtrack,omitempty,flow"`
	Opensecrets    string   `yaml:"opensecrets,omitempty,flow"`
	Votesmart      string   `yaml:"votesmart,omitempty,flow"`
	Fec            []string `yaml:"fec,omitempty,flow"`
	Cspan          string   `yaml:"cspan,omitempty,flow"`
	Wikipedia      string   `yaml:"wikipedia,omitempty,flow"`
	HouseHistory   string   `yaml:"house_history,omitempty,flow"`
	Ballotpedia    string   `yaml:"ballotpedia,omitempty,flow"`
	Maplight       string   `yaml:"maplight,omitempty,flow"`
	Icpsr          string   `yaml:"icpsr,omitempty,flow"`
	Wikidata       string   `yaml:"wikidata,omitempty,flow"`
	GoogleEntityId string   `yaml:"google_entity_id,omitempty,flow"`
	Name           struct {
		First        string `yaml:"first,omitempty,flow"`
		Last         string `yaml:"last,omitempty,flow"`
		OfficialFull string `yaml:"official_full,omitempty,flow"`
	} `yaml:"name,omitempty,flow"`
	Bio struct {
		Birthday string `yaml:"birthday,omitempty,flow"`
		Gender   string `yaml:"gender,omitempty,flow"`
	}
	Terms []struct {
		Type        string `yaml:"type,omitempty,flow"`
		Start       string `yaml:"Start,omitempty,flow"`
		End         string `yaml:"End,omitempty,flow"`
		State       string `yaml:"State,omitempty,flow"`
		District    string `yaml:"District,omitempty,flow"`
		Party       string `yaml:"Party,omitempty,flow"`
		StateRank   string `yaml:"state_rank,omitempty,flow"`
		Url         string `yaml:"url,omitempty,flow"`
		RssUrl      string `yaml:"rss_url,omitempty,flow"`
		ContactForm string `yaml:"contact_form,omitempty,flow"`
		Address     string `yaml:"address,omitempty,flow"`
		Office      string `yaml:"office,omitempty,flow"`
		Phone       string `yaml:"phone,omitempty,flow"`
	}
}

func DownloadLegislatorsYaml() (downloadpath string, err error) {
	MakeTempDir()
	downloadpath = "tmp/legislators.yaml"
	err = DownloadFile(downloadpath, legislatorYamlUrl)
	if err != nil {
		panic(err)
	}
	log.Info().Msgf("Downloaded: %s", legislatorYamlUrl)
	if err != nil {
		panic(err)
	}
	Prepend(downloadpath, "legislators:")
	return
}

func (c *Legislators) ParseLegislatorsYaml(data []byte) error {
	if err := yaml.Unmarshal(data, c); err != nil {
		return err
	}
	// ... Check for elements of the YAML
	return nil
}

func ReadLegislatorsYaml() {
	pathToYaml := "tmp/legislators.yaml"
	yamlFile, err := ioutil.ReadFile(pathToYaml)
	if err != nil {
		log.Error().Msgf("Error reading %s   #%v ", pathToYaml, err)
	}
	var legislators Legislators
	if err := legislators.ParseLegislatorsYaml(yamlFile); err != nil {
		log.Fatal().Err(err)
	}

	//fmt.Printf("%+v\n", legislators)

}
