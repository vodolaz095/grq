golint:
	# installing golint code quality tools and checking, if it can be started
	cd ~ && go get -u golang.org/x/lint/golint
	golint

deps:
	# install all dependencies required for running application
	go version
	go env

	# installing golang dependencies using golang modules
	go mod download # download dependencies
	go mod verify # ensure dependencies are present
	go mod tidy # ensure go.mod is sane

lint:
	gofmt  -w=true -s=true -l=true ./
	golint ./...
	go vet ./...

check: lint
	go test -v -coverprofile=cover.out ./...

test: check

bench:
	go test -test.bench=.* 

consumer:
	go run example/consumer/main.go

publisher:
	go run example/publisher/main.go
