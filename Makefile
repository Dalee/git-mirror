install:
	go get -u github.com/modocache/gover
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

build:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o git-mirror.exe .
	tar -czf ./git-mirror_windows_amd64.tar.gz ./git-mirror.exe ./README.md
	rm -f ./git-mirror.exe

	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o git-mirror .
	tar -czf ./git-mirror_linux_amd64.tar.gz ./git-mirror ./README.md
	rm -f ./git-mirror

	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o git-mirror .
	tar -czf ./git-mirror_darwin_amd64.tar.gz ./git-mirror ./README.md
	rm -f ./git-mirror

.PHONY: install test build format
