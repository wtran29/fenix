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
	@go build -o ../app/fenix.exe ./cmd/cli


## build: builds the command line tool directory
build:
	@go build -o ./dist/fenix.exe ./cmd/cli