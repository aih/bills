package bills

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	sampleFraction    = .05
	wordSamplePath    = path.Join("..", "..", "wordSampleList.json")
	wordSampleListLen = 10000
)

// For each path to a data file, creates a random sample of tokenized words
// of length sampleFraction * number of tokenized words
// Sends the result to a channel
func CollectWordSample(fpath string, wordSampleStorageChannel chan WordSample, wg *sync.WaitGroup) error {
	defer wg.Done()
	var wordSampleItem WordSample
	billCongressTypeNumber := BillNumberFromPath(fpath)
	log.Info().Msgf("Getting a sample of words for: %s\n", billCongressTypeNumber)
	wordSampleItem.BillCongressTypeNumber = billCongressTypeNumber
	// Add an item of the form {billCongressTypeNumber: [list of sampled words]} to the channel
	file, err := os.ReadFile(fpath)
	if err != nil {
		log.Error().Msgf("Error reading document: %s", err)
		return err
	} else {
		fileText := removeXMLRegexCompiled.ReplaceAllString(string(file), " ")
		wordList := CustomTokenize(fileText)
		wordListLen := len(wordList)
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(wordListLen, func(i, j int) { wordList[i], wordList[j] = wordList[j], wordList[i] })
		sliceNum := int(math.Round(sampleFraction * float64(wordListLen)))
		wordSampleItem.WordList = wordList[:sliceNum]
		log.Info().Msgf("Got a sample of %d words for: %s\n", sliceNum, billCongressTypeNumber)
		log.Info().Msgf("Sample words: \n %s", strings.Join(wordList[:3], " "))
		wordSampleStorageChannel <- wordSampleItem
	}
	return nil
}

// Collects a random sample of tokenized words
// for each bill in 'document.xml' files in the 'congress' directory
// Writes the results to the wordSamplePath
func CollectWordSamplesFromBills(pathToCongressDataDir string) (allWords []string) {
	if pathToCongressDataDir == "" {
		pathToCongressDataDir = PathToCongressDataDir
	}
	defer log.Info().Msg("Done collecting word samples")
	log.Info().Msgf("Getting all files in %s.  This may take a while.\n", pathToCongressDataDir)
	documentXMLFiles, _ := ListDocumentXMLFiles(pathToCongressDataDir)
	wordSampleStorageChannel := make(chan WordSample)
	wgWordSample := &sync.WaitGroup{}
	wgWordSample.Add(len(documentXMLFiles))
	go func() {
		wgWordSample.Wait()
		close(wordSampleStorageChannel)
	}()

	for _, fpath := range documentXMLFiles {
		go CollectWordSample(fpath, wordSampleStorageChannel, wgWordSample)
	}

	billCounter := 0
	for wordSampleItem := range wordSampleStorageChannel {
		billCounter++
		// Subsample and save to total list
		log.Info().Msgf("Sampling and storing word sample for %s.\n", wordSampleItem.BillCongressTypeNumber)
		allWords = append(allWords, wordSampleItem.WordList...)
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allWords), func(i, j int) { allWords[i], allWords[j] = allWords[j], allWords[i] })
	log.Info().Msg("Writing word list to file")
	if len(allWords) > 100 {
		log.Info().Msgf("First 100 words: \n %s", strings.Join(allWords[:100], " "))
	} else {
		log.Info().Msgf("Sample words: \n %s", strings.Join(allWords, " "))
	}
	var wordSampleList []string
	if len(allWords) > wordSampleListLen {
		wordSampleList = allWords[:wordSampleListLen]
	} else {
		wordSampleList = allWords
	}
	wordSampleJson, err := json.Marshal(wordSampleList)
	if err != nil {
		log.Error().Msgf("Error converting list to JSON: %s", err)
	} else {
		os.WriteFile(wordSamplePath, wordSampleJson, 0666)
	}
	return allWords

}
