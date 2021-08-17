package bills

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
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
	incorporateThreshold     = .8
	incorporateRatio         = .2
	scoreThreshold           = .1
	nearlyIdenticalThreshold = .8
	similarScoreThreshold    = .1
	minimumTotal             = 150
)

type docMap struct {
	nGramMap map[string]int
	keys     []string
}

type docMaps map[string]*docMap

type CompareItem struct {
	Score        float64
	ScoreOther   float64 // Score of the other bill
	Explanation  string
	ComparedDocs string
}

func getExplanation(scorei, scorej float64, iTotal, jTotal int) string {
	if scorei == 1 && scorej == 1 {
		return "bills-identical"
	}
	//log.Info().Msgf("%d\n", iTotal)
	//log.Info().Msgf("%d\n", jTotal)
	//log.Info().Msgf("%f\n", scorei)
	//log.Info().Msgf("%f\n", scorej)
	//log.Info().Msg("----")

	// minimumTotal avoids small bills being counted as nearly identical
	if ((iTotal > minimumTotal && jTotal > minimumTotal) || ((1 - scorei/scorej) < similarScoreThreshold)) && (scorei > nearlyIdenticalThreshold) && (scorej > nearlyIdenticalThreshold) {
		return "bills-nearly_identical"
	}
	if scorei < scoreThreshold && scorej < scoreThreshold {
		return "bills-unrelated"
	}
	if (scorei > incorporateThreshold) && scorej/scorei < incorporateRatio {
		return "bills-incorporated_by"
	} else if (scorej > incorporateThreshold) && scorei/scorej < incorporateRatio {
		return "bills-incorporates"
	} else {
		return "bills-some_similarity"
	}
}

// Creates ngrams for files in the list of docPaths
// Returns a map with key = docPath and value = docMap
// Each docMap consists of a map of nGrams to the number of occurences, and a list of the nGrams
func makeBillNgrams(docPaths []string) (nGramMaps docMaps, err error) {
	nGramMaps = make(docMaps)
	for i, docpath := range docPaths {
		log.Debug().Msgf("Getting Ngrams for file: %d\n", i)
		file, err := os.ReadFile(docpath)
		if err != nil {
			log.Error().Msgf("Error reading document: %s\n", err)
			return nil, err
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
func compareFiles(nGramMaps docMaps, docPaths []string) (compareMatrix [][]CompareItem, err error) {
	log.Info().Msg("Comparing files")
	compareMatrix = make([][]CompareItem, len(docPaths))
	for i, docpath1 := range docPaths {
		log.Debug().Msgf("Comparison for file: %d\n", i)
		compareMatrix[i] = make([]CompareItem, len(docPaths))
		for j := 0; j < (i + 1); j++ {
			docpath2 := docPaths[j]

			iTotal := 0
			jTotal := 0

			iKeys := nGramMaps[docpath1].keys
			//log.Info().Msg(docpath1)
			//log.Info().Msg(iKeys)
			jKeys := nGramMaps[docpath2].keys
			iScore := 0
			for _, key := range jKeys {
				iValue := nGramMaps[docpath1].nGramMap[key]
				//log.Info().Msg(iValue)
				//jValuea := nGramMaps[docpath2].nGramMap[key]
				//log.Info().Msg(jValuea)
				iScore += iValue
				jTotal += nGramMaps[docpath2].nGramMap[key]
			}
			jScore := 0
			for _, key := range iKeys {
				jValue := nGramMaps[docpath2].nGramMap[key]
				//log.Info().Msg(iValue)
				//jValuea := nGramMaps[docpath2].nGramMap[key]
				//log.Info().Msg(jValuea)
				jScore += jValue
				iTotal += nGramMaps[docpath1].nGramMap[key]
			}

			//if docpath1 != docpath2 {
			//log.Info().Msgf("Keys:\n%v", nGramMaps[docpath1].keys)
			//log.Info().Msgf("Keys 2:\n%v", nGramMaps[docpath2].keys)
			//log.Info().Msgf("scoreTotal: %d\n", scoreTotal)
			//log.Info().Msgf("scoreTotal: %d\n", scoreTotal)
			//log.Info().Msgf("%s vs. %s", docpath1, docpath2)
			//log.Info().Msgf("iTotal: %d\n", iTotal)
			//log.Info().Msgf("jTotal: %d\n", jTotal)
			//}
			scorei := math.Round(100*float64(iScore)/float64(jTotal)) / 100
			scorej := math.Round(100*float64(jScore)/float64(iTotal)) / 100
			exi := getExplanation(scorei, scorej, iTotal, jTotal)
			exj := getExplanation(scorej, scorei, iTotal, jTotal)
			//log.Info().Msgf("i,j docpath1/docpath2 scorei scorej: %d,%d %d/%d %f %f\n", i, j, iTotal, jTotal, scorei, scorej)
			//if exi == "incorporated by" || exj == "incorporated by" {
			//	log.Info().Msgf("i,j docpath1/docpath2 scorei scorej: %d,%d %d/%d %f %f\n", i, j, iTotal, jTotal, scorei, scorej)
			//}

			compareMatrix[i][j] = CompareItem{scorei, scorej, exi, BillNumberFromPath(docpath1) + "-" + BillNumberFromPath(docpath2)}
			compareMatrix[j][i] = CompareItem{scorej, scorei, exj, BillNumberFromPath(docpath2) + "-" + BillNumberFromPath(docpath1)}
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
				log.Info().Msgf("Ticker: %s", t)
			}
		}
	}()

	nGramMaps, err := makeBillNgrams(dOC_PATHS)
	if err != nil {
		log.Panic().Msgf("Error making ngrams: %s\n", err)
	}
	compareMatrix, _ := compareFiles(nGramMaps, dOC_PATHS)
	log.Info().Msg(fmt.Sprint(compareMatrix))
	ticker.Stop()
	done <- true
	log.Info().Msg("Ticker stopped")
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

func CompareBills(parentPath string, billList []string, print bool) ([][]CompareItem, error) {

	var docPathsToCompare []string
	for _, billNumber := range billList {
		billPath, err := PathFromBillNumber(billNumber)
		if err != nil {
			log.Info().Msg("Could not get path for " + billNumber)
		} else {
			// log.Info().Msg(path.Join(parentPath, billPath))
			docPathsToCompare = append(docPathsToCompare, path.Join(parentPath, billPath, "document.xml"))
		}
	}
	// log.Info().Msg(docPathsToCompare)
	nGramMaps, err := makeBillNgrams(docPathsToCompare)
	if err != nil {
		log.Error().Msgf("Error making ngrams: %s\n", err)
		return nil, err
	}
	if len(docPathsToCompare) == 0 {
		log.Info().Msg("No documents to compare")
		if print {
			fmt.Print(":compareMatrix:", "", ":compareMatrix:")
		}
		return nil, nil
	}
	compareMatrix, _ := compareFiles(nGramMaps, docPathsToCompare)
	compareMatrixJson, _ := json.Marshal(compareMatrix)
	if print {
		fmt.Print(":compareMatrix:", string(compareMatrixJson), ":compareMatrix:")
	}
	return compareMatrix, nil
}

func GetCompareMap(compareRow []CompareItem) (compareMap map[string]CompareItem) {
	compareMap = make(map[string]CompareItem)
	log.Debug().Msgf("compareRow: %v", compareRow)
	for _, row := range compareRow {
		comparedocs := strings.Split(row.ComparedDocs, "-")
		if len(comparedocs) == 2 {
			compareMap[comparedocs[1]] = row
		}
	}
	return compareMap
}
