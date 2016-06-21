package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alexcesaro/log/stdlog"
)

// Override the defaults using --url=xxxx and --token=yyyy and -filename=courses.txt
var canvasBase = flag.String("url", "https://vericite.instructure.com/api/v1/", "the base URL for the Canvas API")
var canvasAuth = flag.String("token", "xxxxxx", "the Canvas authentication token after the word Bearer")
var csvFilename = flag.String("filename", "courses.csv", "a file containing all course ids")

// Use -log=debug to get debug-level output
var logger = stdlog.GetFromFlags()

// CanvasAssignment represents an assignment in Canvas
type CanvasAssignment struct {
	AnonymousPeerReviews      bool        `json:"anonymous_peer_reviews"`
	AssignmentGroupID         int         `json:"assignment_group_id"`
	AutomaticPeerReviews      bool        `json:"automatic_peer_reviews"`
	CourseID                  int         `json:"course_id"`
	CreatedAt                 string      `json:"created_at"`
	Description               string      `json:"description"`
	DueAt                     interface{} `json:"due_at"`
	ExternalToolTagAttributes struct {
		NewTab         bool   `json:"new_tab"`
		ResourceLinkID string `json:"resource_link_id"`
		URL            string `json:"url"`
	} `json:"external_tool_tag_attributes"`
	GradeGroupStudentsIndividually bool        `json:"grade_group_students_individually"`
	GradingStandardID              interface{} `json:"grading_standard_id"`
	GradingType                    string      `json:"grading_type"`
	GroupCategoryID                interface{} `json:"group_category_id"`
	HasOverrides                   bool        `json:"has_overrides"`
	HasSubmittedSubmissions        bool        `json:"has_submitted_submissions"`
	HTMLURL                        string      `json:"html_url"`
	ID                             int         `json:"id"`
	IntegrationData                struct{}    `json:"integration_data"`
	IntegrationID                  interface{} `json:"integration_id"`
	LockAt                         interface{} `json:"lock_at"`
	LockedForUser                  bool        `json:"locked_for_user"`
	ModeratedGrading               bool        `json:"moderated_grading"`
	Muted                          bool        `json:"muted"`
	Name                           string      `json:"name"`
	NeedsGradingCount              int         `json:"needs_grading_count"`
	OnlyVisibleToOverrides         bool        `json:"only_visible_to_overrides"`
	PeerReviews                    bool        `json:"peer_reviews"`
	PointsPossible                 int         `json:"points_possible"`
	Position                       int         `json:"position"`
	PostToSis                      interface{} `json:"post_to_sis"`
	Published                      bool        `json:"published"`
	SubmissionTypes                []string    `json:"submission_types"`
	SubmissionsDownloadURL         string      `json:"submissions_download_url"`
	UnlockAt                       interface{} `json:"unlock_at"`
	Unpublishable                  bool        `json:"unpublishable"`
	UpdatedAt                      string      `json:"updated_at"`
	URL                            string      `json:"url"`
}

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

		// Get all assignments inside this course
		req, err := http.NewRequest("GET", *canvasBase+"courses/"+courseID+"/assignments?per_page=100", nil)
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
			logger.Alert("Could not fetch assignments for course. Is your token correct? " + resp.Status)
			return
		}

		// Convert the Canvas JSON into Go struct
		var canvasAssignments []CanvasAssignment
		json.Unmarshal(body, &canvasAssignments)

		logger.Infof("Course "+courseID+" assignment count: %d", len(canvasAssignments))

		// Loop over each assignment and look for the relevant attribute
		for _, canvasAssignment := range canvasAssignments {
			urlToTest := string(canvasAssignment.ExternalToolTagAttributes.URL)
			if strings.Contains(urlToTest, "longsight.com") {
				logger.Debug("Assignment name: " + canvasAssignment.Name + "; URL: " + urlToTest)
				assignmentID := strconv.Itoa(canvasAssignment.ID)

				// Here is the correct VeriCite URL
				data := url.Values{}
				data.Set("assignment[external_tool_tag_attributes][url]", "https://app.vericite.com/vericite/")

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
					logger.Info("Modified assignment: " + assignmentID + ";Return: " + string(resp.Status))
				} else {
					logger.Debug("Request dump: " + string(dump))
					logger.Warning("Request body: " + string(body))
				}
				resp.Body.Close()
				time.Sleep(1 * time.Second)
			}
		}
	}
}
