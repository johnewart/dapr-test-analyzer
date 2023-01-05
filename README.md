
# Dapr test analyzer tool 

Either export an environment variable, or create a .env file with your GitHub token. The name of the variable is `GITHUB_TOKEN` 

Fetch data from GitHub:

`go run ./scripts/fetch.go`

Extract available results: 

`go run ./scripts/extract.go`

Serve the app:

`go build && ./test-analyzer`  or `go run ./scripts/http.go`


Then access `http://localhost:5000`


### TODO

* Simplify things a bit...