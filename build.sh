#!/bin/bash
set -e

APP_NAME="novel-reader"
LDFLAGS="-s -w"

echo "=== 当前平台 ==="
go build -ldflags "$LDFLAGS" -o "${APP_NAME}" .

echo "=== Windows (amd64) ==="
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "${APP_NAME}-windows-amd64.exe" .

echo "=== macOS (amd64) ==="
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "${APP_NAME}-darwin-amd64" .

echo "=== macOS (arm64) ==="
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "${APP_NAME}-darwin-arm64" .

echo "=== 完成 ==="
ls -la ${APP_NAME}* 2>/dev/null || ls -la
