package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

/*
   Extracts messages by a certain person
*/

var folder = "../../downloads"
var from = "Michael Burry"

type messages []string

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

	msgs := messages(fileNames)

	sort.Sort(&msgs)

	fileNames = []string(msgs)

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

var msgNumRex = regexp.MustCompile(`message-(.*)-cleaned`)

func (s *messages) Len() int {
	return len([]string(*s))
}

func (s *messages) Less(i, j int) bool {
	sm := *s

	numStr1 := msgNumRex.FindStringSubmatch(sm[i])[1]
	numStr2 := msgNumRex.FindStringSubmatch(sm[j])[1]

	num1, err := strconv.Atoi(numStr1)
	if err != nil {
		log.Fatalln("Error when converting", numStr1, "to integer")
	}

	num2, err := strconv.Atoi(numStr2)
	if err != nil {
		log.Fatalln("Error when converting", numStr2, "to integer")
	}

	return num1 < num2
}

func (s *messages) Swap(i, j int) {
	sm := *s

	tmp := sm[j]
	sm[j] = sm[i]
	sm[i] = tmp
}

func main() {
	files := readFileNames(folder)
	msgs := extractMessages(files, from)

	ioutil.WriteFile(
		path.Join(folder, from+"-messages"),
		[]byte(msgs), os.ModePerm)
}
