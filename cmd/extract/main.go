package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

/*
   Extracts messages by a certain person
*/

var folder = "../../downloads"
var from = "Michael Burry"

func readFileNames(folder string) []string {
	dir, err := os.Open(folder)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		log.Fatalln("Error reading the folder", folder, err)
	}

	var fileNames []string

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		if !strings.Contains(f.Name(), "cleaned") {
			continue
		}

		fileNames = append(fileNames, path.Join(folder, f.Name()))
	}

	return fileNames
}

func extractMessages(fileNames []string, from string) string {
	var messageExtract = `((?s)%BOM.*?%EOM.*?\n)`
	var messageExtractRex = regexp.MustCompile(messageExtract)
	var result string

	for _, fname := range fileNames {
		log.Println("Reading contents of", fname)
		byteContents, err := ioutil.ReadFile(fname)
		if err != nil {
			log.Println("Error when reading", fname, err)
			continue
		}

		matches := messageExtractRex.FindAllString(string(byteContents), -1)
		if matches == nil || len(matches) == 0 {
			continue
		}

		for _, v := range matches {
			if strings.Contains(v, "From:"+from) {
				result += v
			}
		}
	}

	return result
}

func main() {
	files := readFileNames(folder)
	msgs := extractMessages(files, from)

	ioutil.WriteFile(
		path.Join(folder, from+"-messages"),
		[]byte(msgs), os.ModePerm)
}
