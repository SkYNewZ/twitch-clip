#!/usr/bin/env bash

set -e

GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
TEMP=$(mktemp -d)

echo "Building Go app"
go build -ldflags="-X 'main.twitchClientID=${TWITCH_CLIENT_ID}' -X 'main.twitchClientSecret=${TWITCH_SECRET_ID}'" \
  -tags production \
  -o "${TEMP}/twitch_clip_${GOOS}_${GOARCH}" .

echo "Packaging macOS app"
go run scripts/macapp.go \
  -assets "${TEMP}" \
  -bin "twitch_clip_${GOOS}_${GOARCH}" \
  -icon ./assets/icon1080.png \
  -identifier com.skynewz.twitchclip -name "Twitch Clip" \
  -o ./out/
