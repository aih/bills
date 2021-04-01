package bills

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

var (
	committeesYamlUrl = "https://raw.githubusercontent.com/unitedstates/congress-legislators/master/committees-current.yaml"
)

type Committees struct {
	Committees []Committee
}
type Committee struct {
	Type             string         `yaml:"type,omitempty,flow"`
	Name             string         `yaml:"name,omitempty,flow"`
	Url              string         `yaml:"url,omitempty,flow"`
	MinorityUrl      string         `yaml:"minority_url,omitempty,flow"`
	ThomasId         string         `yaml:"thomas_id,omitempty,flow"`
	HouseCommitteeId string         `yaml:"house_committee_id,omitempty,flow"`
	Subcommittees    []Subcommittee `yaml:"subcommittees,omitempty,flow"`
	Address          string         `yaml:"address,omitempty,flow"`
	Phone            string         `yaml:"phone,omitempty,flow"`
	RssUrl           string         `yaml:"rss_url,omitempty,flow"`
	Jurisdiction     string         `yaml:"jurisdiction,omitempty,flow"`
}

type Subcommittee struct {
	Name     string `yaml:"name,omitempty,flow"`
	ThomasId string `yaml:"thomas_id,omitempty,flow"`
	Address  string `yaml:"address,omitempty,flow"`
	Phone    string `yaml:"phone,omitempty,flow"`
}

func DownloadCommitteesYaml() (downloadpath string, err error) {
	MakeTempDir()
	downloadpath = "tmp/committees.yaml"
	err = DownloadFile(downloadpath, committeesYamlUrl)
	if err != nil {
		panic(err)
	}
	fmt.Println("Downloaded: " + committeesYamlUrl)
	return
}

func (c *Committees) ParseYaml(data []byte) error {
	if err := yaml.Unmarshal(data, c); err != nil {
		return err
	}
	// ... Check for elements of the YAML
	return nil
}

func ReadCommitteesYaml() {
	pathToYaml := "tmp/committees.yaml"
	yamlFile, err := ioutil.ReadFile(pathToYaml)
	if err != nil {
		log.Printf("Error reading %s   #%v ", pathToYaml, err)
	}
	var committees Committees
	if err := committees.ParseYaml(yamlFile); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", committees)

}
