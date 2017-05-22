package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
  "strconv"
	"github.com/alexcesaro/log/stdlog"
)

// Override the defaults using --url=xxxx and --token=yyyy and -filename=courses.txt
var canvasBase = flag.String("url", "https://vericite.instructure.com/api/v1/", "the base URL for the Canvas API")
var canvasAuth = flag.String("token", "xxxxxx", "the Canvas authentication token after the word Bearer")
var csvFilename = flag.String("filename", "assignments.csv", "a file containing all assignment ids")
var outputFolder = flag.String("outputFolder", "submissions", "a path for where the submissions will be stored")
var RESULTS_PER_PAGE = 100
// Use -log=debug to get debug-level output
var logger = stdlog.GetFromFlags()

// CanvasSubmission represents an assignment in Canvas
type CanvasSubmission struct {
	AssignmentId                   int         `json:"assignment_id"`
	Attempt                        int         `json:"attempt"`
	Body                           string      `json:"body"`
	Grade                          string      `json:"grade"`
	GradeMatchesCurrentSubmission  bool        `json:"grade_matches_current_submission"`
  HtmlUrl                        string      `json:"html_url"`
	PreviewUrl                     string      `json:"preview_url"`
	Score                          float32     `json:"score"`
	SubmissionType                 string      `json:"submission_type"`
	SubmittedAt                    string      `json:"submitted_at"`
	URL                            string      `json:"url"`
	UserId                         int         `json:"user_id"`
	GraderId                       int         `json:"grader_id"`
	Late                           bool        `json:"late"`
	Excused                        bool        `json:"excused"`
	WorkflowState                  string      `json:"workflow_state"`
	Attachments []struct {
		Id         int   `json:"id"`
		FileName   string `json:"filename"`
		URL            string `json:"url"`
	} `json:"attachments"`
}

func main() {
	client := &http.Client{}

	file, err := os.Open(*csvFilename)
	if err != nil {
		panic("Cannot open CSV. Please supply a valid path to a CSV file.")
	}
	defer file.Close()
	reader := csv.NewReader(file)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic("Problem reading file")
		}
		courseID := record[0]
		assignmentID := record[1]
		if _, err := strconv.Atoi(courseID); err != nil {
			//this is most likely the header, skip
      continue
		}

		// Get all assignments inside this course
		var page = 1
		for {
			req, err := http.NewRequest("GET", *canvasBase+"courses/"+courseID+"/assignments/"+assignmentID+"/submissions?per_page=" + strconv.Itoa(RESULTS_PER_PAGE) + "&page=" + strconv.Itoa(page), nil)
			fmt.Println(*canvasBase+"courses/"+courseID+"/assignments/"+assignmentID+"/submissions?per_page=" + strconv.Itoa(RESULTS_PER_PAGE) + "&page=" + strconv.Itoa(page))
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
				logger.Warning("Could not fetch submissions for course " + courseID + " assignment: " + assignmentID + ". Canvas response: " + resp.Status)
				break
			}

			// Convert the Canvas JSON into Go struct
			var canvasSubmissions []CanvasSubmission
			json.Unmarshal(body, &canvasSubmissions)

			// Loop over each assignment and look for the relevant attribute
			for _, canvasSubmission := range canvasSubmissions {
				if(canvasSubmission.SubmissionType == "online_upload" && len(canvasSubmission.Attachments) > 0){
					for _, attachment := range canvasSubmission.Attachments {
						if len(attachment.URL) > 0 {
							downloadFromUrl(attachment.URL, *outputFolder + "/" + courseID + "/" + assignmentID, strconv.Itoa(attachment.Id) + attachment.FileName)
						}
					}
				}
			}
			if(len(canvasSubmissions) >= RESULTS_PER_PAGE && page < 100){ //limit results to 100 * RESULTS_PER_PAGE
				//more results, go to next page:
				page++
			}else{
				//no more results, break out of for Loop
				break;
			}
		}
	}
}

func downloadFromUrl(url string, filePath string, fileName string) {
	fmt.Println(filePath + "/" + fileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
    err = os.MkdirAll(filePath, 0755)
    if err != nil {
      panic(err)
    }
  }

	output, err := os.Create(filePath + "/" + fileName)
	if err != nil {
		fmt.Println("Error while creating", fileName, "-", err)
		return
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}

	fmt.Println(n, "bytes downloaded.")
}
