#!/bin/bash
# if any error detected, stop build
set -e
# assume rasberryPi arm V8 processor
GOARCH_ARG="${1:-arm64}"
# get absolute path to project directory assuming running build.sh from build directory
coffeePath="$( dirname "$( pwd )")"
# run compile
docker run --rm -v "$coffeePath":/go/src/CoffeeM -w /go/src/CoffeeM/coffeem  golang:1.9.2 /go/src/CoffeeM/build/build_go.sh "${GOARCH_ARG}" 
# put static binary coffeem in its own container
echo 'INFO: Creating docker container'
docker build -t coffeem  ../coffeem >/dev/null
echo "INFO: Created docker container: 'coffeem:latest'"

