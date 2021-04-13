package bills

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	if len(documentXMLFiles) == 0 {
		fmt.Printf("Retrying on %s\n", PathToDataDir)
		documentXMLFiles, err = WalkDirFilter(PathToDataDir, isDocumentXML)
		if err == nil {
			fmt.Printf("Got %d files!\n", len(documentXMLFiles))
		}
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

// Make local tmp directory if it doesn't exist
func MakeTempDir() {
	fmt.Println("Making tmp directory")
	path := "./tmp"
	mode := os.ModePerm
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, mode)
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
		fmt.Println(err)
		return err
	}
	_, err = tmpcopy.WriteString(text + "\n")
	if err != nil {
		fmt.Println(err)
		tmpcopy.Close()
		return err
	}
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}
	_, err = tmpcopy.Write(data)
	if err != nil {
		log.Fatal(err)
	}

	e := CopyFile(tmpname, filepath)
	if e != nil {
		log.Fatal(e)
	}

	return nil
}
