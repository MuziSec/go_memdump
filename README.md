# go_memdump
Really bad Go Wrapper for collecting memory dumps and chunking into smaller files. (CS Falcon has issues with files larger than 4 GB, so chunking is required to pull back via Falcon)

## Compile

For Windows:
```
GOOS=windows GOARCH=amd64 go build main.go
```
For Linux:
```
GOOS=linux GOARCH=amd64 go build main.go
```
