#!/usr/bin/env bash

set -e

VERSION=$(git describe --tags --exact-match 2>/dev/null || git describe --tags 2>/dev/null || echo "v0.0.0-$(git rev-parse --short HEAD)")
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
TEMP=$(mktemp -d)

echo "Building Go app"
go build -ldflags="-X 'main.twitchClientID=${TWITCH_CLIENT_ID}' -X 'main.twitchClientSecret=${TWITCH_SECRET_ID}'" \
  -tags production \
  -o "${TEMP}/twitch_clip_${VERSION}_${GOOS}_${GOARCH}" .

echo "Packaging macOS app"
go run scripts/macapp.go \
  -assets "${TEMP}" \
  -bin "twitch_clip_${VERSION}_${GOOS}_${GOARCH}" \
  -icon ./assets/icon1080.png \
  -identifier com.skynewz.twitchclip -name "Twitch Clip" \
  -o ./out/
