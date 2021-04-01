package bills

import (
	"fmt"
)

var (
	legislatorYamlUrl = "https://raw.githubusercontent.com/unitedstates/congress-legislators/master/legislators-current.yaml"
)

func DownloadLegislatorsYaml() (downloadpath string, err error) {
	MakeTempDir()
	downloadpath = "tmp/legislators.yaml"
	err = DownloadFile(downloadpath, legislatorYamlUrl)
	if err != nil {
		panic(err)
	}
	fmt.Println("Downloaded: " + legislatorYamlUrl)
	return
}
