## Synopsis

This script uses the Canvas API to adjust the assignment field "external_tool_tag_attributes" to correct the VeriCite URL.

## Script Options

```
 -filename string
        a file containing all course ids (default "courses.csv")
  -log string
        sets the logging threshold (default "info")
  -token string
        the Canvas authentication token after the word Bearer (default "xxxxxx")
  -url string
        the base URL for the Canvas API (example "https://acmecollege.instructure.com/api/v1/")
```

## Running

```
wget https://github.com/vericite/canvas-utils/raw/master/rewrite-assignment-urls
chmod +x rewrite-assignment-urls
./rewrite-assignment-urls -token="9000~aXXXXXXXXXXXXXXXXXXX" -url="https://acmecollege.instructure.com/api/v1/"
```

## Building from Source

```
git clone https://github.com/vericite/canvas-utils.git
cd canvas-utils
go build rewrite-assignment-urls.go
```

## Debugging

Add -log=debug to command above

## License

MIT
