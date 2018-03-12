#!/bin/bash
set -e
echo "INFO: Downloading dependencies"
go get -u github.com/golang/dep/cmd/dep
dep ensure
echo "INFO: Dependencies finished"
export GOOS=linux
export GOARCH="$1" 
echo "INFO: Compiling go for GOOS: '$GOOS' GOARCH: '$GOARCH'"
go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o coffeem 
echo "INFO: Compiling finished"


