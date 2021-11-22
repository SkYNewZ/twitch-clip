#!/usr/bin/env bash

set -e

GO="go.exe"
VERSION=$(git describe --tags --exact-match 2>/dev/null || git describe --tags 2>/dev/null || echo "v0.0.0-$(git rev-parse --short HEAD)")
GOOS=$($GO env GOOS)
GOARCH=$($GO env GOARCH)

echo "Installing go-winres"
$GO install github.com/tc-hib/go-winres@latest

echo "Generating Windows manifests"
export VERSION
go generate ./...

echo "Building Go app"
$GO build -ldflags="-s -w -X 'main.twitchClientID=${TWITCH_CLIENT_ID}' -X 'main.twitchClientSecret=${TWITCH_SECRET_ID}' -H=windowsgui" \
  -tags production \
  -o "out/twitch_clip_${GOOS}_${GOARCH}.exe" .
