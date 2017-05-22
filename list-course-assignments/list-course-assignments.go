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
var csvFilename = flag.String("filename", "courses.csv", "a file containing all course ids")
var turnitin = flag.Bool("turnitin", false, "A flag indicating to only return assignments with TurnItIn enabled")
var RESULTS_PER_PAGE = 100
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
	TurnitinEnabled                bool        `json:"turnitin_enabled"`
}

func main() {
	client := &http.Client{}

	file, err := os.Open(*csvFilename)
	if err != nil {
		panic("Cannot open CSV. Please supply a valid path to a CSV file.")
	}
	defer file.Close()
	reader := csv.NewReader(file)

	// Loop through the file containing course IDs
	fmt.Printf("courseId,assignmentId,assignmentName\n")
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic("Problem reading file")
		}
		courseID := record[0]
		if _, err := strconv.Atoi(courseID); err != nil {
			//this is most likely the header, skip
      continue
		}

		// Get all assignments inside this course
		var page = 1
		for {
			req, err := http.NewRequest("GET", *canvasBase+"courses/"+courseID+"/assignments?per_page=" + strconv.Itoa(RESULTS_PER_PAGE) + "&page=" + strconv.Itoa(page), nil)
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
				logger.Warning("Could not fetch assignments for course " + courseID + ". Canvas response: " + resp.Status)
				break
			}

			// Convert the Canvas JSON into Go struct
			var canvasAssignments []CanvasAssignment
			json.Unmarshal(body, &canvasAssignments)

			// Loop over each assignment and look for the relevant attribute
			for _, canvasAssignment := range canvasAssignments {
				if(((len(canvasAssignment.SubmissionTypes) == 2 && contains(canvasAssignment.SubmissionTypes, "online_upload") && contains(canvasAssignment.SubmissionTypes, "online_text_entry")) ||
					 (len(canvasAssignment.SubmissionTypes) == 1 && contains(canvasAssignment.SubmissionTypes, "online_upload")) ||
					 (len(canvasAssignment.SubmissionTypes) == 1 && contains(canvasAssignment.SubmissionTypes, "online_text_entry"))) &&
					 (*turnitin != true || canvasAssignment.TurnitinEnabled == true)){
					 fmt.Printf("%v,%v,%v\n", courseID, canvasAssignment.ID, canvasAssignment.Name)
				}
			}
			if(len(canvasAssignments) >= RESULTS_PER_PAGE && page < 100){ //limit results to 100 * RESULTS_PER_PAGE
				//more results, go to next page:
				page++
			}else{
				//no more results, break out of for Loop
				break;
			}
		}
	}
}

func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}
