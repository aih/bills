package bills

import (
	"fmt"
)

var (
	committeesYamlUrl = "https://raw.githubusercontent.com/unitedstates/congress-legislators/master/committees-current.yaml"
)

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
