#!/bin/bash
./Config.sh
coffeePath="$( dirname "$( pwd )")"
docker run --rm -v "$coffeePath":/go/src/CoffeeM -w /go/src/CoffeeM/coffeem -e GOOS=linux -e GOARCH=amd64 golang:1.9.2 go build -v -a -tags netgo -ldflags '-w -extldflags "-static"' -o coffeem 
docker build -t coffeem  ../coffeem

