#!/bin/bash
Config.sh
docker run --rm -v "../coffeem":/usr/src/myapp -w /usr/src/myapp -e GOOS=linux -e GOARCH=armv64 golang:1.9.2 go build -v
