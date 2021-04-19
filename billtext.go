package bills

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path"
	"sync"
	"time"
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
	fmt.Printf("Getting a sample of words for: %s\n", billCongressTypeNumber)
	wordSampleItem.BillCongressTypeNumber = billCongressTypeNumber
	// Add an item of the form {billCongressTypeNumber: [list of sampled words]} to the channel
	file, err := os.ReadFile(fpath)
	if err != nil {
		log.Printf("Error reading document: %s", err)
		return err
	} else {
		fileText := removeXMLRegexCompiled.ReplaceAllString(string(file), " ")
		wordList := CustomTokenize(fileText)
		wordListLen := len(wordList)
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(wordListLen, func(i, j int) { wordList[i], wordList[j] = wordList[j], wordList[i] })
		sliceNum := int(math.Round(sampleFraction * float64(wordListLen)))
		wordSampleItem.WordList = wordList[:sliceNum]
		fmt.Printf("Got a sample of %d words for: %s\n", sliceNum, billCongressTypeNumber)
		fmt.Println("Sample words: \n", wordList[:3])
		wordSampleStorageChannel <- wordSampleItem
	}
	return nil
}

// Collects a random sample of tokenized words
// for each bill in 'document.xml' files in the 'congress' directory
// Writes the results to the wordSamplePath
func CollectWordSamplesFromBills(pathToCongressDataDir string) {
	if pathToCongressDataDir == "" {
		pathToCongressDataDir = PathToCongressDataDir
	}
	defer fmt.Println("Done collecting word samples")
	fmt.Printf("Getting all files in %s.  This may take a while.\n", pathToCongressDataDir)
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
	allWords := make([]string, 0)
	for wordSampleItem := range wordSampleStorageChannel {
		billCounter++
		// Subsample and save to total list
		fmt.Printf("Sampling and storing word sample for %s.\n", wordSampleItem.BillCongressTypeNumber)
		allWords = append(allWords, wordSampleItem.WordList...)
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allWords), func(i, j int) { allWords[i], allWords[j] = allWords[j], allWords[i] })
	fmt.Println("Writing word list to file")
	fmt.Println("First 100 words: \n", allWords[:100])
	wordSampleList := make([]string, 0)
	if len(allWords) > wordSampleListLen {
		wordSampleList = allWords[:wordSampleListLen]
	} else {
		wordSampleList = allWords
	}
	wordSampleJson, err := json.Marshal(wordSampleList)
	if err != nil {
		log.Printf("Error converting list to JSON: %s", err)
	} else {
		os.WriteFile(wordSamplePath, wordSampleJson, 0666)
	}

}
