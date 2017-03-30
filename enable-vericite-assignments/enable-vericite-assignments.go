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
	"time"

	"github.com/alexcesaro/log/stdlog"
)

// Override the defaults using --url=xxxx and --token=yyyy and -filename=courses.txt
var canvasBase = flag.String("url", "https://vericite.instructure.com/api/v1/", "the base URL for the Canvas API")
var canvasAuth = flag.String("token", "xxxxxx", "the Canvas authentication token after the word Bearer")
var csvFilename = flag.String("filename", "assignments.csv", "a file containing all course ids")
var visibility = flag.String("visibility", "immediate", "Option: Students Can See the Originality Report")
var exclude_quoted = flag.String("excludeQuoted", "true", "Option: Exclude Quoted Material")
var exclude_self_plag = flag.String("excludeSelfPlag", "true", "Option: Exclude Self Plagiarism")
var store_in_index = flag.String("storeInIndex", "true", "Option: Store submissions in Institutional Index")
// var uploadEntry = flag.String("uploadEntry", "true", "Option: Upload entry setting")
// var textEntry = flag.String("textEntry", "true", "Option: Text entry setting")

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
	VericiteEnabled                bool        `json:"vericite_enabled"`
	VeriCiteSettings               struct {
    OriginalityReportVisibility         string   `json:"originality_report_visibility"`
		ExcludeQuotes                       bool     `json:"exclude_quoted"`
		ExcludeSelfPlag                     bool     `json:"exclude_self_plag"`
		StoreInIndex                        bool     `json:"store_in_index"`
	} `json:"turnitin_settings"`
}

func main() {
	client := &http.Client{}

	file, err := os.Open(*csvFilename)
	if err != nil {
		panic("Can not open CSV")
	}
	defer file.Close()
	reader := csv.NewReader(file)

	//validate parameters:
	if(*visibility != "immediate" &&
			*visibility != "after_grading" &&
			*visibility != "after_due_date" &&
			*visibility != "never"){
			panic("Visibility parameter can only be one of the following: immediate, after_grading, after_due_date, never")
	}
	if(*exclude_quoted != "true" && *exclude_quoted != "false"){
		panic("excludeQuoted can only be true or false")
	}
	if(*exclude_self_plag != "true" && *exclude_self_plag != "false"){
		panic("excludeSelfPlag can only be true or false")
	}
	if(*store_in_index != "true" && *store_in_index != "false"){
		panic("storeInIndex can only be true or false")
	}
	// if(*textEntry != "true" && *textEntry != "false"){
	// 	panic("textEntry can only be true or false")
	// }
	// if(*uploadEntry != "true" && *uploadEntry != "false"){
	// 	panic("uploadEntry can only be true or false")
	// }
	// if(*uploadEntry == "false" && *textEntry == "false"){
	// 	panic("Either textEntry or uploadEntry must be true")
	// }
	// Loop through the file containing course IDs
	logger.Info("VeriCite settings:\nVisibility: " + *visibility + "\nExcludeQuotes: " + *exclude_quoted + "\nExclude Self Plag: " + *exclude_self_plag + "\nStore in Index: " + *store_in_index)
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
			//courseId is not a number, skip
			continue
		}

		// Here is the correct VeriCite URL
		data := url.Values{}
		data.Set("assignment[turnitin_enabled]", "false")
		data.Set("assignment[vericite_enabled]", "true")
		data.Set("assignment[turnitin_settings][originality_report_visibility]", *visibility)
		data.Set("assignment[turnitin_settings][exclude_quoted]", *exclude_quoted)
		data.Set("assignment[turnitin_settings][exclude_self_plag]", *exclude_self_plag)
		data.Set("assignment[turnitin_settings][store_in_index]", *store_in_index)
		var encodedParams = data.Encode()
		// if(*uploadEntry == "true"){
		// 	encodedParams = "assignment%5Bsubmission_types%5D%5B%5D=online_upload&" + encodedParams
		// }
		// if(*textEntry == "true"){
		// 	encodedParams = "assignment%5Bsubmission_types%5D%5B%5D=online_text_entry&" + encodedParams
		// }

		// Create an HTTP PUT to modify this one assignment field
		r, _ := http.NewRequest("PUT", *canvasBase+"courses/"+courseID+"/assignments/"+assignmentID,
			bytes.NewBufferString(encodedParams))
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
			logger.Info("Modified assignment: " + assignmentID + ";Canvas response: " + string(resp.Status))
		} else {
			logger.Debug("Request dump: " + string(dump))
			logger.Warning("Request body: " + string(body))
		}
		resp.Body.Close()
		time.Sleep(1 * time.Second)
	}
}
