.PHONY: target 

COMPILE_TIME = $(shell date +"%Y-%m-%d-%H%M%S")
TAG = $(shell git describe)

target:
	export CGO_ENABLED=0 && \
	export GOOS=linux && \
	export GOARCH=amd64 && \
	go build -ldflags '-w -s' -o lambdaHandler .
