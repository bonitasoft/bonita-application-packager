#!/usr/bin/env bash

PACKAGE_NAME="bonita"

version="${1:-development}"

base_command=(CGO_ENABLED=0 go build -ldflags="-X 'github.com/bonitasoft/bonita-application-packager/cmd.Version=${version}'")

echo "Building binary '${version}' for Windows..."
env GOOS=windows GOARCH=amd64 "${base_command[@]}" -o dist/windows-amd64/${PACKAGE_NAME}.exe
if [ $? -ne 0 ]; then
    echo 'An error has occurred building binary for Windows! Aborting the script execution...'
    exit 1
fi

echo "Building binary '${version}' for Linux..."
env GOOS=linux GOARCH=amd64 "${base_command[@]}" -o dist/linux-amd64/${PACKAGE_NAME}
if [ $? -ne 0 ]; then
    echo 'An error has occurred building binary for Linux! Aborting the script execution...'
    exit 1
fi

echo "Building binary '${version}' for MacOS amd64..."
env GOOS=darwin GOARCH=amd64 "${base_command[@]}" -o dist/darwin-amd64/${PACKAGE_NAME}
if [ $? -ne 0 ]; then
    echo 'An error has occurred building binary for MacOS amd64! Aborting the script execution...'
    exit 1
fi

echo "Building binary '${version}' for MacOS arm64..."
env GOOS=darwin GOARCH=arm64 "${base_command[@]}" -o dist/darwin-arm64/${PACKAGE_NAME}
if [ $? -ne 0 ]; then
    echo 'An error has occurred building binary for MacOS arm64! Aborting the script execution...'
    exit 1
fi
