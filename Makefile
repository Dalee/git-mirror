install:
	go get -u github.com/modocache/gover
	go get -u github.com/golang/lint/golint
	go get -u github.com/Masterminds/glide
	go get -u github.com/gordonklaus/ineffassign
	go get -u github.com/client9/misspell/cmd/misspell

test:
	golint -set_exit_status .
	ineffassign ./
	misspell -error README.md .
	gofmt -d -s -e .

format:
	gofmt -d -w -s -e .

build-linux:
	GOOS=linux GOARCH=amd64 go build -o git-mirror git-mirror.go config.go

.PHONY: test build-linux format
