package bills

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
)

type FilterFunc func(string) bool

// Walk directory with a filter. Returns the filepaths that
// pass the 'testPath' function
func WalkDirFilter(root string, testPath FilterFunc) (filePaths []string, err error) {
	defer fmt.Println("Done collecting filepaths.")
	fmt.Printf("Getting all file paths in %s.  This may take a while.\n", root)
	filePaths = make([]string, 0)
	accumulate := func(fpath string, entry fs.DirEntry, err error) error {
		if err != nil {
			log.Println(err)
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
func ListDocumentXMLFiles() (documentXMLFiles []string, err error) {
	isDocumentXML := func(fpath string) bool {
		_, file := filepath.Split(fpath)
		return file == "document.xml"
	}
	documentXMLFiles, err = WalkDirFilter(PathToCongressDataDir, isDocumentXML)
	if err == nil {
		fmt.Printf("Got %d files!\n", len(documentXMLFiles))
	}
	return
}

// Walk 'congress' directory and get filepaths to 'data.json'
// which contains metadata for the bill
func ListDataJsonFiles() (dataJsonFiles []string, err error) {
	isDataJson := func(fpath string) bool {
		_, file := filepath.Split(fpath)
		return file == "data.json"
	}
	dataJsonFiles, err = WalkDirFilter(PathToCongressDataDir, isDataJson)
	if err == nil {
		fmt.Printf("Got %d files!\n", len(dataJsonFiles))
	}
	return
}
