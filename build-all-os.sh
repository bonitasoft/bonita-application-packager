#!/usr/bin/env bash

echo 'Building binary for Windows...'
env GOOS=windows GOARCH=amd64 go build -o dist/windows-amd64/
if [ $? -ne 0 ]; then
    echo 'An error has occurred building binary for Windows! Aborting the script execution...'
    exit 1
fi
echo 'Done!'

echo 'Building binary for Linux...'
env GOOS=linux GOARCH=amd64 go build -o dist/linux-amd64/
if [ $? -ne 0 ]; then
    echo 'An error has occurred building binary for Linux! Aborting the script execution...'
    exit 1
fi
echo 'Done!'

echo 'Building binary for MacOS amd64...'
env GOOS=darwin GOARCH=amd64 go build -o dist/darwin-amd64/
if [ $? -ne 0 ]; then
    echo 'An error has occurred building binary for MacOS amd64! Aborting the script execution...'
    exit 1
fi
echo 'Done!'

echo 'Building binary for MacOS arm64...'
env GOOS=darwin GOARCH=arm64 go build -o dist/darwin-arm64/
if [ $? -ne 0 ]; then
    echo 'An error has occurred building binary for MacOS arm64! Aborting the script execution...'
    exit 1
fi
echo 'Done!'
