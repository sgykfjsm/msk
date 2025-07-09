BINARY_NAME := msk
VERSION := v0.1.0
COMMIT_HASH := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -s -w \
		  -X main.Version=$(VERSION) \
		  -X main.CommitHash=$(COMMIT_HASH) \
		  -X main.BuildDate=$(BUILD_DATE)

.PHONY: dependencies
dependencies:
	go mod tidy
	go mod verify

.PHONY: download
download:
	go mod download

.PHONY: build
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .

.PHONY: clean
clean:
	rm -f $(BINARY_NAME)

# Run commands
.PHONY: clusterinfo
clusterinfo:
	./$(BINARY_NAME) clusterinfo

.PHONY: generate-notice
generate-notice:
	./$(BINARY_NAME) generate-notice

.PHONY: notify
notify:
	./$(BINARY_NAME) notify
