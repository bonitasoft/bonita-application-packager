#!/usr/bin/env bash

PACKAGE_NAME="bonita-packager"

version="${1:-development}"

echo "Building binary ${version} for Windows..."
env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o dist/windows-amd64/${PACKAGE_NAME}.exe -ldflags="-X 'main.Version=${version}'"
if [ $? -ne 0 ]; then
    echo 'An error has occurred building binary for Windows! Aborting the script execution...'
    exit 1
fi

echo "Building binary ${version} for Linux..."
env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/linux-amd64/${PACKAGE_NAME} -ldflags="-X 'main.Version=${version}'"
if [ $? -ne 0 ]; then
    echo 'An error has occurred building binary for Linux! Aborting the script execution...'
    exit 1
fi

echo "Building binary ${version} for MacOS amd64..."
env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o dist/darwin-amd64/${PACKAGE_NAME} -ldflags="-X 'main.Version=${version}'"
if [ $? -ne 0 ]; then
    echo 'An error has occurred building binary for MacOS amd64! Aborting the script execution...'
    exit 1
fi

echo "Building binary ${version} for MacOS arm64..."
env GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o dist/darwin-arm64/${PACKAGE_NAME} -ldflags="-X 'main.Version=${version}'"
if [ $? -ne 0 ]; then
    echo 'An error has occurred building binary for MacOS arm64! Aborting the script execution...'
    exit 1
fi
