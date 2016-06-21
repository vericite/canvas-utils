## Synopsis

This script uses the Canvas API to adjust the assignment field "external_tool_tag_attributes" to correct the VeriCite URL.

## Script Options

 -filename string
        a file containing all course ids (default "courses.csv")
  -log string
        sets the logging threshold (default "info")
  -token string
        the Canvas authentication token after the word Bearer (default "xxxxxx")
  -url string
        the base URL for the Canvas API (example "https://acmecollege.instructure.com/api/v1/")

## Running

./rewrite-assignment-urls -token="9000~aXXXXXXXXXXXXXXXXXXX" -url="https://acmecollege.instructure.com/api/v1/"

## Building

go build -x rewrite-assignment-urls.go

## License

MIT
