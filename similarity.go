package bills

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"time"
)

var (
	pATH_TO_CONGRESSDATA_DIR_116_HR = path.Join(PathToCongressDataDir, "data", "116", "bills", "hr")
	pATH_TO_CONGRESSDATA_DIR_116_S  = path.Join(PathToCongressDataDir, "data", "116", "bills", "s")
	iH_PATH                         = path.Join("text-versions", "ih", "document.xml")
	iS_PATH                         = path.Join("text-versions", "is", "document.xml")
	rH_PATH                         = path.Join("text-versions", "rh", "document.xml")
	eNR_PATH                        = path.Join("text-versions", "enr", "document.xml")
	dOC_PATHS                       = []string{
		path.Join(pATH_TO_CONGRESSDATA_DIR_116_HR, "hr203", iH_PATH),  // similar
		path.Join(pATH_TO_CONGRESSDATA_DIR_116_HR, "hr299", iH_PATH),  // similar
		path.Join(pATH_TO_CONGRESSDATA_DIR_116_S, "s1195", iS_PATH),   // similar
		path.Join(pATH_TO_CONGRESSDATA_DIR_116_HR, "hr1500", iH_PATH), // not similar
		path.Join(pATH_TO_CONGRESSDATA_DIR_116_HR, "hr700", iH_PATH),  // not similar
		path.Join(pATH_TO_CONGRESSDATA_DIR_116_HR, "hr133", eNR_PATH), // incorporates 7617
		path.Join(pATH_TO_CONGRESSDATA_DIR_116_HR, "hr7617", rH_PATH), // incorporated in 133
	}
	incorporateThreshold  = .8
	incorporateRatio      = .2
	scoreThreshold        = .1
	similarThreshold      = .8
	similarScoreThreshold = .1
	minimumTotal          = 150
)

type docMap struct {
	nGramMap map[string]int
	keys     []string
}

type docMaps map[string]*docMap

type CompareItem struct {
	Score       float64
	Explanation string
}

func getExplanation(scorei, scorej float64, iTotal, jTotal int) string {
	if scorei == 1 && scorej == 1 {
		return "_identical_"
	}
	//fmt.Println(iTotal)
	//fmt.Println(jTotal)
	//fmt.Println(scorei)
	//fmt.Println(scorej)
	//fmt.Println("----")

	// minimumTotal avoids small bills being counted as nearly identical
	if ((iTotal > minimumTotal && jTotal > minimumTotal) || ((1 - scorei/scorej) < similarScoreThreshold)) && scorei > similarThreshold && scorej > similarThreshold {
		return "_nearly_identical_"
	}
	if scorei < scoreThreshold && scorej < scoreThreshold {
		return "_unrelated_"
	}
	if (scorei > incorporateThreshold) && scorej/scorei < incorporateRatio {
		return "_incorporated_by_"
	} else if (scorej > incorporateThreshold) && scorei/scorej < incorporateRatio {
		return "_incorporates_"
	} else {
		return "_some_similarity_"
	}
}

// Creates ngrams for files in the list of docPaths
// Returns a map with key = docPath and value = docMap
// Each docMap consists of a map of nGrams to the number of occurences, and a list of the nGrams
func makeBillNgrams(docPaths []string) (nGramMaps docMaps, err error) {
	nGramMaps = make(docMaps)
	for i, docpath := range docPaths {
		fmt.Printf("Getting Ngrams for file: %d\n", i)
		file, err := os.ReadFile(docpath)
		if err != nil {
			log.Printf("Error reading document: %s\n", err)
			return nGramMaps, err
		} else {
			fileText := removeXMLRegexCompiled.ReplaceAllString(string(file), " ")
			var docMapItem *docMap = new(docMap)
			docMapItem.nGramMap = MakeNgramMap(fileText, 4)
			docMapItem.keys = MapNgramKeys(docMapItem.nGramMap)

			nGramMaps[docpath] = docMapItem
		}
	}
	return nGramMaps, nil

}

// Compares all of the documents in a docMaps object, returns a matrix of the comparison values
func compareFiles(nGramMaps docMaps) (compareMatrix [][]CompareItem, err error) {
	fmt.Println("Comparing files")
	docPaths := make([]string, len(nGramMaps))
	compareMatrix = make([][]CompareItem, len(docPaths))
	d := 0
	for docpath := range nGramMaps {
		docPaths[d] = docpath
		d++
	}
	for i, docpath1 := range docPaths {
		fmt.Printf("Comparison for file: %d\n", i)
		compareMatrix[i] = make([]CompareItem, len(docPaths))
		for j := 0; j < (i + 1); j++ {
			docpath2 := docPaths[j]

			iTotal := 0
			jTotal := 0

			iKeys := nGramMaps[docpath1].keys
			//fmt.Println(docpath1)
			//fmt.Println(iKeys)
			jKeys := nGramMaps[docpath2].keys
			iScore := 0
			for _, key := range jKeys {
				iValue := nGramMaps[docpath1].nGramMap[key]
				//fmt.Println(iValue)
				//jValuea := nGramMaps[docpath2].nGramMap[key]
				//fmt.Println(jValuea)
				iScore += iValue
				jTotal += nGramMaps[docpath2].nGramMap[key]
			}
			jScore := 0
			for _, key := range iKeys {
				jValue := nGramMaps[docpath2].nGramMap[key]
				//fmt.Println(iValue)
				//jValuea := nGramMaps[docpath2].nGramMap[key]
				//fmt.Println(jValuea)
				jScore += jValue
				iTotal += nGramMaps[docpath1].nGramMap[key]
			}

			//if docpath1 != docpath2 {
			// fmt.Printf("Keys:\n%v", nGramMaps[docpath1].keys)
			// fmt.Printf("Keys 2:\n%v", nGramMaps[docpath2].keys)
			// fmt.Printf("scoreTotal: %d\n", scoreTotal)
			// fmt.Printf("scoreTotal: %d\n", scoreTotal)
			// fmt.Printf("iTotal: %d\n", iTotal)
			// fmt.Printf("jTotal: %d\n", jTotal)
			//}
			scorei := math.Round(100*float64(iScore)/float64(jTotal)) / 100
			scorej := math.Round(100*float64(jScore)/float64(iTotal)) / 100
			exi := getExplanation(scorei, scorej, iTotal, jTotal)
			exj := getExplanation(scorej, scorei, iTotal, jTotal)
			//if exi == "incorporated by" || exj == "incorporated by" {
			//	fmt.Printf("i,j docpath1/docpath2 scorei scorej: %d,%d %d/%d %f %f\n", i, j, iTotal, jTotal, scorei, scorej)
			//}

			compareMatrix[i][j] = CompareItem{scorei, exi}
			compareMatrix[j][i] = CompareItem{scorej, exj}
		}
	}

	return compareMatrix, nil
}

// Compares a sample list of documents, defined in dOC_PATHS
func CompareSamples() {
	ticker := time.NewTicker(500 * time.Millisecond)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Ticker: ", t)
			}
		}
	}()

	nGramMaps, _ := makeBillNgrams(dOC_PATHS)
	compareMatrix, _ := compareFiles(nGramMaps)
	fmt.Println(compareMatrix)
	ticker.Stop()
	done <- true
	fmt.Println("Ticker stopped")
	// NOTE if [i][j] > [j][i] then i is incorporated in j and j incorporates i
	// Returns:
	//[
	//	[{1 identical} {0.95 nearly identical} {0.3 incorporates} {0.01 unrelated} {0.01 unrelated} {0.11 incorporated by} {0.03 unrelated}]
	//	[{0.91 nearly identical} {1 identical} {0.29 incorporates} {0.01 unrelated} {0.01 unrelated} {0.1 incorporated by} {0.02 unrelated}]
	//	[{0.82 incorporated by} {0.82 incorporated by} {1 identical} {0.03 unrelated} {0.03 unrelated} {0.13 incorporated by} {0.04 unrelated}]
	//	[{0.01 unrelated} {0.01 unrelated} {0.01 unrelated} {1 identical} {0.01 unrelated} {0.04 unrelated} {0.01 unrelated}]
	//	[{0.07 unrelated} {0.07 unrelated} {0.05 unrelated} {0.06 unrelated} {1 identical} {0.06 unrelated} {0.04 unrelated}]
	//	[{0 incorporates} {0 incorporates} {0 incorporates} {0 unrelated} {0 unrelated} {1 identical} {0.03 incorporates}]
	//	[{0 unrelated} {0 unrelated} {0 unrelated} {0 unrelated} {0 unrelated} {0.84 incorporated by} {1 identical}]
	//	]

}

// To call from Python
// import subprocess
// result = subprocess.run(['./compare', '-p', '../../../congress/data', '-b', '116hr1500rh,115hr6972ih'],  capture_output=True, text=True)
// result.stdout.split('compareMatrix:\n')[-1]
// Out[4]: '[[{1 identical} {0.63 incorporates}] [{0.79 incorporated by} {1 identical}]]'

func CompareBills(parentPath string, billList []string) [][]CompareItem {

	var docPathsToCompare []string
	for _, billNumber := range billList {
		billPath, err := PathFromBillNumber(billNumber)
		if err != nil {
			fmt.Println("Could not get path for " + billNumber)
		} else {
			// fmt.Println(path.Join(parentPath, billPath))
			docPathsToCompare = append(docPathsToCompare, path.Join(parentPath, billPath, "document.xml"))
		}
	}
	// fmt.Println(docPathsToCompare)
	nGramMaps, _ := makeBillNgrams(docPathsToCompare)
	compareMatrix, _ := compareFiles(nGramMaps)
	compareMatrixJson, _ := json.Marshal(compareMatrix)
	fmt.Print(":compareMatrix:", string(compareMatrixJson), ":compareMatrix:")
	//fmt.Print("compareMatrix:", compareMatrix, ":compareMatrix")
	return compareMatrix
}
