.PHONY: test build all linux osx

all: osx linux

test:
	go test . ./app

linux:
	go test . ./app
	GOARCH=amd64 GOOS=linux go build
	mv kubemrr ./releases/linux/amd64

osx:
	go test . ./app
	GOARCH=amd64 GOOS=darwin go build
	mv kubemrr ./releases/darwin/amd64
