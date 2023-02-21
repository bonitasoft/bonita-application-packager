#!/usr/bin/env bash

package=bonita-packager

env GOOS=windows GOARCH=amd64 go build -o dist/${package}-amd64.exe
if [ $? -ne 0 ]; then
    echo 'An error has occurred building binary for Windows! Aborting the script execution...'
    exit 1
fi
env GOOS=linux GOARCH=amd64 go build -o dist/${package}-linux
if [ $? -ne 0 ]; then
    echo 'An error has occurred building binary for Linux! Aborting the script execution...'
    exit 1
fi
env GOOS=darwin GOARCH=amd64 go build -o dist/${package}-amd64-darwin
if [ $? -ne 0 ]; then
    echo 'An error has occurred building binary for MacOS amd64! Aborting the script execution...'
    exit 1
fi
env GOOS=darwin GOARCH=arm64 go build -o dist/${package}-arm64-darwin
if [ $? -ne 0 ]; then
    echo 'An error has occurred building binary for MacOS arm64! Aborting the script execution...'
    exit 1
fi
