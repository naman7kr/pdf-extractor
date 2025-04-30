#!/bin/bash

mkdir -p outputs/mac outputs/linux outputs/windows

GODS=darwin GOARCH=amd64 go build -o outputs/mac/pdf-extractor main.go
GODS=linux GOARCH=amd64 go build -o outputs/linux/pdf-extractor main.go
GODS=windows GOARCH=amd64 go build -o outputs/windows/pdf-extractor.exe main.go

echo "Build complete for macOS, linux and windows"
