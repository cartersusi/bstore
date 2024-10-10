GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

BINARY_NAME=bstore

LDFLAGS=-ldflags="-s -w"
GCFLAGS=-gcflags="-m -l -B"

PLATFORMS=linux darwin windows
ARCHITECTURES=amd64 arm64

all: build

build:
	$(GOBUILD) $(LDFLAGS) $(GCFLAGS) -o $(BINARY_NAME) .

prod:
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHITECTURES); do \
			echo "Building for $$platform/$$arch..."; \
			mkdir -p prod; \
			GOOS=$$platform GOARCH=$$arch $(GOBUILD) $(LDFLAGS) $(GCFLAGS) -o prod/$$platform-$$arch .; \
		done; \
	done

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf *-*/

run: build
	./$(BINARY_NAME)

deps:
	$(GOGET) -v -t -d ./...

.PHONY: all build build-prod clean run deps