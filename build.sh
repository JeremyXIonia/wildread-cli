#!/bin/bash
set -e

LDFLAGS="-s -w"

mkdir -p bin

echo "=== 当前平台 ==="
go build -ldflags "$LDFLAGS" -o reader .

echo "=== Windows (amd64) ==="
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o bin/reader-windows-amd64.exe .

echo "=== macOS (amd64 Intel) ==="
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o bin/reader-darwin-amd64 .

echo "=== macOS (arm64 Apple Silicon) ==="
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o bin/reader-darwin-arm64 .

echo "=== 完成 ==="
ls -la bin/reader-*
