BINARY_NAME=fenixApp

## build: builds all binaries
build:
	go mod vendor
	go build -o tmp/${BINARY_NAME} .
	echo Fenix built!

run: build
	echo Starting Fenix...
	.\tmp\${BINARY_NAME} &
	echo Fenix started!

clean:
	echo Cleaning...
	go clean
	del .\tmp\${BINARY_NAME}
	echo Cleaned!

test:
	echo Testing...
	go test ./...
	echo Done!

start: run
    
stop:
	echo "Starting the front end..."
	taskkill /IM ${BINARY_NAME} /F
	echo Stopped Fenix

restart: stop start