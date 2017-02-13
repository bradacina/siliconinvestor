package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var subjectURLTemplate = "http://www.siliconinvestor.com/subject.aspx?subjectid=%v"
var msgURLTemplate = "http://www.siliconinvestor.com/readmsgs.aspx?subjectid=%v&msgNum=%v&batchsize=100&batchtype=Next"
var startMsg = 0
var endMsg = 50000
var subjectID = "10036"
var outputFolder = "../../downloads"
var delayBetwenReq = time.Second * 10

var numPostsRegex = `<a title=['"]Jump to posts['"].*?>(\d+)</a>`
var msgBodyRegex = `(?s)To:.*?<td align=['"]right['"]>(.*?)</td>.*?From:.*?href=['"]profile[.]aspx.*?>(.*?)</a>.*?<span id=['"]intelliTXT['"]>(.*?)</span>`
var hrefRegex = `<a.*?href=['"](.*?)['"].*?</a>`

var urlRexp = regexp.MustCompile(hrefRegex)
var bodyRex = regexp.MustCompile(msgBodyRegex)

var endOfMessage = "%EOM----------------------\n\n"
var beginOfMessage = "%BOM---------------------\n"

func main() {
	os.Remove(outputFolder)

	err := os.MkdirAll(outputFolder, os.ModeDir)
	if err != nil {
		log.Fatal("Error when trying to create outputFolder", outputFolder, err)
	}

	numMsg, err := getNumMessages(subjectID)
	if err != nil {
		return
	}

	numRequests := (numMsg / 100) + 1
	log.Println("We'll need", numRequests, "requests to retrieve all messages")

	for i := startMsg; i < numMsg; i += 100 {
		if i > endMsg {
			break
		}
		download(i)
	}
}

func getNumMessages(subjectid string) (int, error) {
	url := fmt.Sprintf(subjectURLTemplate, subjectid)
	log.Println("Trying to retrieve", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("Error when trying to get number of total messages for subjectId", subjectid, err)
		return 0, err
	}

	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)

	rxp := regexp.MustCompile(numPostsRegex)

	var lines string

	for {
		lineBytes, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		lines += string(lineBytes)
	}

	//log.Println(lines)
	matches := rxp.FindStringSubmatch(lines)

	if len(matches) == 2 {
		log.Println("Total number of messages is", matches[1])
		result, err := strconv.Atoi(matches[1])
		if err != nil {
			log.Println("Error when converting", matches[1], "to int")
			return 0, nil
		}

		return result, nil
	}

	return 0, nil
}

func download(msgNum int) error {
	log.Println("Downloading message", msgNum)
	filename := path.Join(outputFolder, "subjectId-"+subjectID+"-message-"+strconv.Itoa(msgNum))

	url := fmt.Sprintf(msgURLTemplate, subjectID, msgNum)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error when retrieving message", msgNum, err)
		return nil
	}
	defer resp.Body.Close()

	f, err := os.Create(filename)
	if err != nil {
		log.Println("Error when creating file", filename)
		return nil
	}
	defer f.Close()

	var content string

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("Error when reading line from response for message number", msgNum, err)
		}

		content += line
	}

	matches := bodyRex.FindAllStringSubmatch(content, -1)
	if matches == nil || len(matches) == 0 {
		log.Println("Could not find any messages when downloading message number", msgNum)
		return nil
	}

	for _, v := range matches {
		cleaned := v[3]
		cleaned = urlRexp.ReplaceAllString(cleaned, " $1 ")
		cleaned = strings.Replace(cleaned, "<br>", "\n", -1)
		f.WriteString(beginOfMessage)
		f.WriteString("Date:" + v[1] + "\n")
		f.WriteString("From:" + v[2] + "\n")
		f.WriteString(cleaned + "\n")
		f.WriteString(endOfMessage)
	}

	return nil
}
