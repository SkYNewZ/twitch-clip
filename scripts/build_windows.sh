#!/usr/bin/env bash

set -e

VERSION=$(git describe --tags --exact-match 2>/dev/null || git describe --tags 2>/dev/null || echo "v0.0.0-$(git rev-parse --short HEAD)")
export VERSION

echo "Generating Windows manifests"
go generate ./...

echo "Building Go app"
go.exe build -ldflags="-X 'main.twitchClientID=${TWITCH_CLIENT_ID}' -X 'main.twitchClientSecret=${TWITCH_SECRET_ID}' -H=windowsgui" -tags production -o out/twitch-clip.exe .
