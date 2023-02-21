#!/usr/bin/env bash

package=$1
if [[ -z "$package" ]]; then
  echo "usage: $0 <package-name>"
  echo "example: $0 bonita-packager"
  exit 1
fi

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
