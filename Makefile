deps:
	# install all dependencies required for running application
	go version
	go env

	# installing golint code quality tools and checking, if it can be started
	cd ~ && go get -u golang.org/x/lint/golint
	golint

	# installing golang dependencies using golang modules
	go mod tidy # ensure go.mod is sane
	go mod verify # ensure dependencies are present

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
