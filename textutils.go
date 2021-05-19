package bills

import (
	"regexp"
	"sort"
	"strings"

	"github.com/jdkato/prose/tokenize"
)

var (
	puncts = "-./(),!@#$%^&*:\\;"
)

// Tokenizer function that returns words longer than 3 characters
// which do not have certain punctuation. Currently:  "-./(),!@#$%^&*:\\;"
func CustomTokenize(text string) (wordList []string) {
	wordListAll := tokenize.TextToWords(text)
	for _, word := range wordListAll {
		word = strings.TrimSpace(word)
		if len(word) > 3 && !(strings.ContainsAny(word, puncts)) {
			wordList = append(wordList, word)
		}
	}
	return
}

// Creates a map with ngrams as keys and number of occurences as values
// n is the number of words in each n-gram
func MakeNgramMap(text string, n int) (wordMap map[string]int) {
	wordListAll := CustomTokenize(text)
	nGramLen := len(wordListAll) - n
	nGrams := make([]string, nGramLen)
	for i := 0; i < nGramLen; i++ {
		nGrams[i] = strings.Join(wordListAll[i:i+n], " ")
	}
	wordMap = make(map[string]int)
	for _, nGram := range nGrams {
		wordMap[nGram] = wordMap[nGram] + 1
	}
	return
}

// Creates a list of ngrams.
// First makes a map with 'MakeNgramMap'
// Then returns a list of the keys of the map
func MakeNgrams(text string, n int) (wordList []string) {
	nGramMap := MakeNgramMap(text, n)
	wordList = MapNgramKeys(nGramMap)
	return

}

// Returns the keys of a map of type map[string]int
func MapNgramKeys(nGramMap map[string]int) (keys []string) {

	keys = make([]string, len(nGramMap))

	i := 0
	for k := range nGramMap {
		keys[i] = k
		i++
	}
	return
}

var REASON_ORDER = map[string]int{"bills-identical": 1, "bills-nearly_identical": 2, "bills-title_match": 3, "bills-includes": 4, "bills-included_by": 5, "related": 6, "bills-some_similarity": 7, "bills-unrelated": 8}

func SortReasons(reasons []string) []string {

	sort.Slice(reasons, func(i, j int) bool {
		return REASON_ORDER[reasons[i]] < REASON_ORDER[reasons[j]]
	})
	return reasons
}

// Removes duplicates in a list of strings
// Returns the deduplicated list
func RemoveDuplicates(elements []string) []string { // change string to int here if required
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{} // change string to int here if required
	result := []string{}             // change string to int here if required

	for v := range elements {
		if encountered[elements[v]] {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

// Returns a map of regex capture groups to the items that are matched
func FindNamedMatches(regex *regexp.Regexp, str string) map[string]string {
	match := regex.FindStringSubmatch(str)

	results := map[string]string{}
	for i, name := range match {
		results[regex.SubexpNames()[i]] = name
	}
	return results
}
