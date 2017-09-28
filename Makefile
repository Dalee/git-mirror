install:
	go get -u github.com/golang/lint/golint
	go get -u github.com/gordonklaus/ineffassign
	go get -u github.com/client9/misspell/cmd/misspell

test:
	golint -set_exit_status .
	ineffassign ./
	misspell -error README.md .
	gofmt -d -s -e .

format:
	gofmt -d -w -s -e .

.PHONY: install test format
