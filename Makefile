.PHONY: test build all linux osx

release: osx linux

test:
	go test . ./app

linux: test
	GOARCH=amd64 GOOS=linux go build
	mv kubemrr ./releases/linux/amd64

osx: test
	GOARCH=amd64 GOOS=darwin go build
	mv kubemrr ./releases/darwin/amd64
