.PHONY: test build all linux osx

release: set-version osx linux
	git commit -am "set version to $(VERSION)"
	git tag v$(VERSION)
	git push
	git push --tags

test:
	go test . ./app

linux: test
	GOARCH=amd64 GOOS=linux go build
	mv kubemrr ./releases/linux/amd64

osx: test
	GOARCH=amd64 GOOS=darwin go build
	mv kubemrr ./releases/darwin/amd64

set-version:
ifndef VERSION
	$(error VERSION is not set)
endif
	if ! git diff-index --quiet HEAD ; then echo "you have uncommitted changes"; exit 1 ; fi
	sed -i s:'VERSION = "[^"]*"':'VERSION = "$(VERSION)"':g app/version.go
	sed -i s:/raw/v[^/]*:/raw/v$(VERSION):g README.md
