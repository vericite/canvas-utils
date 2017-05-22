package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"github.com/alexcesaro/log/stdlog"
)

// Override the defaults using --url=xxxx and --token=yyyy and -filename=courses.txt
var canvasBase = flag.String("url", "https://vericite.instructure.com/api/v1/", "the base URL for the Canvas API")
var canvasAuth = flag.String("token", "xxxxxx", "the Canvas authentication token after the word Bearer")
var csvFilename = flag.String("filename", "assignments.csv", "a file containing all assignment ids")

// Use -log=debug to get debug-level output
var logger = stdlog.GetFromFlags()

func main() {
	client := &http.Client{}

	file, err := os.Open(*csvFilename)
	if err != nil {
		panic("Can not open CSV")
	}
	defer file.Close()
	reader := csv.NewReader(file)

	// Loop through the file containing course IDs
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic("Problem reading file")
		}
		courseID := record[0]
		assignmentID := record[1]
		assignmentName := record[2]
		if _, err := strconv.Atoi(courseID); err != nil {
			//this is most likely the header, skip
      continue
		}

		data := url.Values{}
		data.Set("assignment[external_tool_tag_attributes][url]", "https://api.vericite.com/web/v1/authenticate/lti")
		// Create an HTTP PUT to modify this one assignment field
		r, _ := http.NewRequest("PUT", *canvasBase+"courses/"+courseID+"/assignments/"+assignmentID,
			bytes.NewBufferString(data.Encode()))
		r.Header.Add("Authorization", "Bearer "+*canvasAuth)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

		resp, err := client.Do(r)
		if err != nil {
			panic("Could not do request")
		}
		dump, _ := httputil.DumpRequestOut(r, true)
		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode <= 206 {
			logger.Info("Modified assignment: " + courseID + ":" + assignmentID + ":" + assignmentName + ";Canvas response: " + string(resp.Status))
		} else {
			logger.Debug("Request dump: " + string(dump))
			logger.Warning("Request body: " + string(body))
		}
		resp.Body.Close()
	}
}
