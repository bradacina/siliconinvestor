package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"unicode"
)

var spaceReplace = regexp.MustCompile(`[\s]+`)
var emptyLinesReplace = regexp.MustCompile(`\n\n\n+`)

var folder = "../downloads"
var splitAfter = 79 // how many chars per line max

func readFileNames() []string {
	dir, err := os.Open(folder)
	if err != nil {
		log.Fatalln("Could not open the folder", folder, err)
	}

	files, err := dir.Readdir(-1)
	if err != nil {
		log.Fatalln("Error reading the folder", folder, err)
	}

	var fileNames []string

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		if strings.Contains(f.Name(), "cleaned") {
			continue
		}

		fileNames = append(fileNames, path.Join(folder, f.Name()))
	}

	return fileNames
}

func readFile(name string) string {
	log.Println("Reading file", name)
	f, err := os.Open(name)
	if err != nil {
		log.Println("Could not open file", name, err)
		return ""
	}
	defer f.Close()

	var cleaned string

	rdr := bufio.NewReader(f)

	for {
		line, err := rdr.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		line = strings.Trim(line, " \t\n\r")
		line = strings.Replace(line, "<br>", "\n", -1)
		line = strings.Replace(line, "\t", " ", -1)
		line = strings.Replace(line, "&#39;", "'", -1)
		line = strings.Replace(line, "&amp;", "&", -1)
		line = strings.Replace(line, "&gt;", ">", -1)
		line = strings.Replace(line, "&lt;", "<", -1)
		line = strings.Replace(line, "<b>", "", -1)
		line = strings.Replace(line, "</b>", "", -1)
		line = strings.Replace(line, "<i>", "", -1)
		line = strings.Replace(line, "</i>", "", -1)

		line = spaceReplace.ReplaceAllString(line, " ")

		lines := splitIntoLines(line)

		for _, l := range lines {
			cleaned += l + "\n"
		}
	}

	cleaned = emptyLinesReplace.ReplaceAllString(cleaned, "\n\n")
	return cleaned
}

func splitIntoLines(line string) []string {
	var ret []string

	runes := []rune(line)

	if len(runes) > splitAfter {
		for len(runes) > splitAfter {
			splitAt := splitAfter
			for {
				if unicode.IsSpace(runes[splitAt]) {
					break
				}
				splitAt--

				// we didn't find any space to split at
				if splitAt == 0 {
					splitAt = splitAfter
					break
				}
			}

			ret = append(ret, string(runes[:splitAt]))
			runes = runes[splitAt+1:]
		}

		if len(runes) > 0 {
			ret = append(ret, string(runes))
		}

	} else {
		ret = append(ret, line)
	}

	return ret
}

func writeFile(name, content string) {
	f, err := os.Create(name)
	if err != nil {
		log.Println("Could not create file", name, err)
		return
	}

	defer f.Close()

	n, err := f.WriteString(content)
	if err != nil {
		log.Println("Error when trying to write to file", name, err)
		return
	}
	if n != len(content) {
		log.Println("Error: not all bytes were written to file", name)
	}
}

func main() {
	files := readFileNames()

	for _, name := range files {
		cleaned := readFile(name)
		writeFile(name+"-cleaned", cleaned)
	}
}
