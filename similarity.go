package bills

import (
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
	incorporateThreshold = .5
	scoreThreshold       = .1
	similarThreshold     = .9
)

type docMap struct {
	nGramMap map[string]int
	keys     []string
}

type docMaps map[string]*docMap

type compareItem struct {
	score       float64
	explanation string
}

func getExplanation(scorei, scorej float64) string {
	if scorei == 1 && scorej == 1 {
		return "_identical"
	}
	if scorei > similarThreshold && scorej > similarThreshold {
		return "_nearly_identical"
	}
	if scorei < scoreThreshold && scorej < scoreThreshold {
		return "_unrelated"
	}
	if (scorei > incorporateThreshold) && scorei > scorej {
		return "_incorporated by"
	} else if (scorej > incorporateThreshold) && scorej > scorei {
		return "_incorporates"
	} else {
		return "_some_similarity"
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
func compareFiles(nGramMaps docMaps) (compareMatrix [][]compareItem, err error) {
	fmt.Println("Comparing files")
	docPaths := make([]string, len(nGramMaps))
	compareMatrix = make([][]compareItem, len(docPaths))
	d := 0
	for docpath := range nGramMaps {
		docPaths[d] = docpath
		d++
	}
	for i, docpath1 := range docPaths {
		fmt.Printf("Comparison for file: %d\n", i)
		compareMatrix[i] = make([]compareItem, len(docPaths))
		for j := 0; j < (i + 1); j++ {
			docpath2 := docPaths[j]

			iTotal := 0
			jTotal := 0

			allKeys := RemoveDuplicates(append(nGramMaps[docpath1].keys, nGramMaps[docpath2].keys...))
			scoreTotal := 0
			for _, key := range allKeys {
				iValue := nGramMaps[docpath1].nGramMap[key]
				jValue := nGramMaps[docpath2].nGramMap[key]
				iTotal = iTotal + iValue
				jTotal = jTotal + jValue
				score := int(math.Min(float64(iValue), float64(jValue)))
				scoreTotal = scoreTotal + score
			}
			//if docpath1 != docpath2 {
			//	fmt.Printf("Keys:\n%v", nGramMaps[docpath1].keys)
			//	fmt.Printf("Keys 2:\n%v", nGramMaps[docpath2].keys)
			//		fmt.Printf("scoreTotal: %d\n", scoreTotal)
			//		fmt.Printf("scoreTotal: %d\n", scoreTotal)
			//		fmt.Printf("iTotal: %d\n", iTotal)
			//		fmt.Printf("jTotal: %d\n", jTotal)
			//}
			scorei := math.Round(100*float64(scoreTotal)/float64(iTotal)) / 100
			scorej := math.Round(100*float64(scoreTotal)/float64(jTotal)) / 100
			exi := getExplanation(scorei, scorej)
			exj := getExplanation(scorej, scorei)
			//if exi == "incorporated by" || exj == "incorporated by" {
			//	fmt.Printf("i,j docpath1/docpath2 scorei scorej: %d,%d %d/%d %f %f\n", i, j, iTotal, jTotal, scorei, scorej)
			//}
			newItemi := compareItem{scorei, exi}
			newItemj := compareItem{scorej, exj}

			compareMatrix[i][j] = newItemi
			compareMatrix[j][i] = newItemj
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

func CompareBills(parentPath string, billList []string) [][]compareItem {

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
	fmt.Print("compareMatrix:", compareMatrix, ":compareMatrix")
	return compareMatrix
}
