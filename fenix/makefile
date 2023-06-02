## Test: runs all tests
test: 
	@go test -v ./...

## cover: opens coverage in browser
cover:
	@go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

## coverage: displays test coverage
coverage:
	@go test -cover ./...

## build_cli: builds the command line tool Fenix and copies it to the app
build_cli:
	@go build -o ../app/Fenix ./cmd/cli