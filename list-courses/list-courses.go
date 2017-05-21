package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/alexcesaro/log/stdlog"
)

// Override the defaults using --url=xxxx and --token=yyyy and -filename=courses.txt
var canvasBase = flag.String("url", "https://vericite.instructure.com/api/v1/", "the base URL for the Canvas API")
var canvasAuth = flag.String("token", "xxxxxx", "the Canvas authentication token after the word Bearer")
var accountId = flag.String("accountId", "", "account id to look up courses")
var termId = flag.String("termId", "", "term id for requested account courses")

// Use -log=debug to get debug-level output
var logger = stdlog.GetFromFlags()

// CanvasCourse represents an assignment in Canvas
type CanvasCourse struct {
	ID                        int         `json:"id"`
	Name                      string      `json:"name"`
	AccountID                 int         `json:"account_id"`
}

func main() {
	client := &http.Client{}
	fmt.Printf("id,name\n")

	// Get all assignments inside this course
	req, err := http.NewRequest("GET", *canvasBase+"accounts/"+*accountId+"/courses?per_page=100&enrollment_term_id="+*termId, nil)
	if err != nil {
		panic("Could not fetch: " + *canvasBase + "courses")
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+*canvasAuth)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("Auth failed fetching")
	}

	if resp.StatusCode != http.StatusOK {
		logger.Warning("Could not fetch courses for account " + *accountId + " and term: " + *termId + ". Canvas response: " + resp.Status)
	}else{
		// Convert the Canvas JSON into Go struct
		var canvasCourses []CanvasCourse
		json.Unmarshal(body, &canvasCourses)

		// Loop over each assignment and look for the relevant attribute
		for _, canvasCourse := range canvasCourses {
			fmt.Printf("%v,%v\n", canvasCourse.ID, canvasCourse.Name)
		}
	}
}
