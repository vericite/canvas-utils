# SCRIPT: list-courses

This script uses the Canvas API to print out a list of the Courses associated with the given Canvas Account ID and Term ID parameters. You will want to save the output into a CSV file named courses.csv to use for other scripts input.

### Script Options

```
  -token string (required)
        the Canvas authentication token after the word Bearer (default "xxxxxx")
  -url string (required)
        the base URL for the Canvas API (example "https://acmecollege.instructure.com/api/v1/")
  -accountId (required)
        the Canvas Account Id that you wish to list courses for
  -termId (required)
        the Canvas Term Id that you wish to list courses for
```

### Example
```
./list-courses -token="9000~aXXXXXXXXXXXXXXXXXXX" -url="https://acmecollege.instructure.com/api/v1/" -accountId=1 -termId=1 > courses.csv
```

# SCRIPT: list-course-assignments

This script uses the Canvas API to print out a list of the Assignments associated with the courses.csv input file. It will only print out assignments that have a submission type of "online_upload" or "online_text_entry" or both. You will want to save the output into a CSV file named assignments.csv to use for other scripts input.

### Script Options

```
 -filename string (required)
        a file containing all course ids
  -log string
        sets the logging threshold (default "info")
  -token string (required)
        the Canvas authentication token after the word Bearer (default "xxxxxx")
  -url string (required)
        the base URL for the Canvas API (example "https://acmecollege.instructure.com/api/v1/")
  -turnitin (optional)
        optional flag to only return assignments that have turnitin enabled (turnitin must still be enabled in Canvas for this to work)
  -vericiteLtiMigration (optional)
        optional flag to only return assignments that have an old VeriCite LTI URL
```

### Example
```
./list-course-assignments -token="9000~aXXXXXXXXXXXXXXXXXXX" -url="https://acmecollege.instructure.com/api/v1/" -filename="courses.csv" > assignments.csv
```

# SCRIPT: export-submissions

This script uses the Canvas API to download submission attachments from a list of assignments into a folder specified by the outputFolder parameter.

### Script Options

```
  -filename string (required)
        a file containing all assignment ids
  -token string (required)
        the Canvas authentication token after the word Bearer (default "xxxxxx")
  -url string (required)
        the base URL for the Canvas API (example "https://acmecollege.instructure.com/api/v1/")
  -outputFolder (default submissions)
        the location where you want to download submissions
```

### Example
```
./export-submissions -token="9000~aXXXXXXXXXXXXXXXXXXX" -url="https://acmecollege.instructure.com/api/v1/" -filename="assignments.csv" -outputFolder="submissions"
```

# SCRIPT: enable-vericite-assignments

This script uses the Canvas API to enable VeriCite for each assignment listed in the assignments.csv input file (CSV with courseId, assignmentId).

### Script Options

```
 -filename string
        a file containing all course and assignment ids (default "assignments.csv")
  -log string
        sets the logging threshold (default "info")
  -token string
        the Canvas authentication token after the word Bearer (default "xxxxxx")
  -url string
        the base URL for the Canvas API (example "https://acmecollege.instructure.com/api/v1/")
  -visibility string (default immediate)
        Option: Students Can See the Originality Report: immediate, after_grading, after_due_date, never
  -excludeQuoted bool (default true)
        Option: Exclude Quoted Material
  -excludeSelfPlag bool (default true)
        Option: Exclude Self Plagiarism
  -storeInIndex bool (default true)
        Option: Store submissions in Institutional Index
```

### Example
CSV File (output from list-course-assignments script, only courseId and assignmentId are used):
```
courseId,assignmentId,assignmentName
1,33,VeriCite Internal 1
1,34,VeriCite LTI
1,45,VC Local LTI 2
```
Run Script:
```
./enable-vericite-assignments -token="9000~aXXXXXXXXXXXXXXXXXXX" -url="https://acmecollege.instructure.com/api/v1/" -filename="assignments.csv"
```

# SCRIPT: rewrite-assignment-urls

This script uses the Canvas API to adjust the assignment field "external_tool_tag_attributes" to correct the VeriCite URL. It is used as a migration from a previous LTI (external tool) URL to a new one. It takes the input of the "list-course-assignments" script (make sure to set the "vericiteLtiMigration" flag to true).

### Script Options

```
 -filename string
        a file containing all course ids (default "assignments.csv")
  -log string
        sets the logging threshold (default "info")
  -token string
        the Canvas authentication token after the word Bearer (default "xxxxxx")
  -url string
        the base URL for the Canvas API (example "https://acmecollege.instructure.com/api/v1/")
```

### Example
```
./rewrite-assignment-urls -token="9000~aXXXXXXXXXXXXXXXXXXX" -url="https://acmecollege.instructure.com/api/v1/"
```

# Combine scripts in a chain of output and input

The scripts are written so that you can combine them

Ex: Download all submissions for a given account ID and term ID:

```
./list-courses -token="9000~aXXXXXXXXXXXXXXXXXXX" -url="https://acmecollege.instructure.com/api/v1/" -accountId=1 -termId=1 | tee courses.csv \
&& ./list-course-assignments -token="9000~aXXXXXXXXXXXXXXXXXXX" -url="https://acmecollege.instructure.com/api/v1/" -filename="courses.csv" | tee assignments.csv \
&& ./export-submissions -token="9000~aXXXXXXXXXXXXXXXXXXX" -url="https://acmecollege.instructure.com/api/v1/" -filename="assignments.csv" -outputFolder="submissions"
```


# Building from Source
Example shown for rewrite-assignment-url
```
git clone https://github.com/vericite/canvas-utils.git
cd canvas-utils/rewrite-assignment-urls
go build rewrite-assignment-urls.go
```

# Cross-compilation
Example shown for rewrite-assignment-url

Build a Windows version from Linux

```
GOOS=windows GOARCH=386 go build -o rewrite-assignment-urls.exe rewrite-assignment-urls.go
```

Build a Mac version from Linux

```
GOOS=darwin go build -o mac-rewrite-assignment-urls rewrite-assignment-urls.go
```

# Debugging

Add -log=debug to command above

# License

MIT
