package bills

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

type FilterFunc func(string) bool

// Walk directory with a filter. Returns the filepaths that
// pass the 'testPath' function
func WalkDirFilter(root string, testPath FilterFunc) (filePaths []string, err error) {
	defer log.Info().Msg("Done collecting filepaths.")
	log.Info().Msgf("Getting all file paths in %s.  This may take a while.\n", root)
	filePaths = make([]string, 0)
	accumulate := func(fpath string, entry fs.DirEntry, err error) error {
		if err != nil {
			log.Error().Err(err)
			return err
		}
		if testPath(fpath) {
			filePaths = append(filePaths, fpath)
		}
		return nil
	}
	err = filepath.WalkDir(root, accumulate)
	return
}

// Walk 'congress' directory and get filepaths to 'document.xml'
// which contains the bill xml
func ListDocumentXMLFiles(pathToCongressDataDir string) (documentXMLFiles []string, err error) {
	if pathToCongressDataDir == "" {
		pathToCongressDataDir = PathToCongressDataDir
	}
	isDocumentXML := func(fpath string) bool {
		_, file := filepath.Split(fpath)
		return file == "document.xml"
	}
	documentXMLFiles, err = WalkDirFilter(pathToCongressDataDir, isDocumentXML)
	if err == nil {
		log.Info().Msgf("Got %d files!\n", len(documentXMLFiles))
	}
	if len(documentXMLFiles) == 0 {
		log.Info().Msgf("Retrying on %s\n", PathToDataDir)
		documentXMLFiles, err = WalkDirFilter(PathToDataDir, isDocumentXML)
		if err == nil {
			log.Info().Msgf("Got %d files!\n", len(documentXMLFiles))
		}
	}
	return
}

// Walk 'congress' directory and get filepaths to 'data.json'
// which contains metadata for the bill
func ListDataJsonFiles(pathToCongressDataDir string) (dataJsonFiles []string, err error) {
	if pathToCongressDataDir == "" {
		pathToCongressDataDir = PathToCongressDataDir
	}
	isDataJson := func(fpath string) bool {
		_, file := filepath.Split(fpath)
		return file == "data.json"
	}
	dataJsonFiles, err = WalkDirFilter(pathToCongressDataDir, isDataJson)
	if err == nil {
		log.Info().Msgf("Got %d files!\n", len(dataJsonFiles))
	}
	if len(dataJsonFiles) == 0 {
		log.Info().Msgf("Retrying on %s\n", PathToDataDir)
		dataJsonFiles, err = WalkDirFilter(PathToDataDir, isDataJson)
		if err == nil {
			log.Info().Msgf("Got %d files!\n", len(dataJsonFiles))
		}
	}
	return
}

// Make local tmp directory if it doesn't exist
func MakeTempDir() {
	log.Info().Msg("Making tmp directory")
	tmpPath := "./tmp"
	mode := os.ModePerm
	if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
		os.Mkdir(tmpPath, mode)
	}
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
// See https://golangcode.com/download-a-file-from-a-url/
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s: %s", resp.Status, resp.Status)
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// Copy the src file to dst. Any existing file will be overwritten and will not
// copy file attributes.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func Prepend(filepath string, text string) error {
	tmpname := "tmpcopy.txt"
	tmpcopy, err := os.Create(tmpname)
	defer os.Remove(tmpname)
	if err != nil {
		log.Error().Err(err)
		return err
	}
	_, err = tmpcopy.WriteString(text + "\n")
	if err != nil {
		log.Error().Err(err)
		tmpcopy.Close()
		return err
	}
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal().Err(err)
	}
	_, err = tmpcopy.Write(data)
	if err != nil {
		log.Fatal().Err(err)
	}

	e := CopyFile(tmpname, filepath)
	if e != nil {
		log.Fatal().Err(e)
	}

	return nil
}
