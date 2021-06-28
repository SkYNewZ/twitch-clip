#!/usr/bin/env bash

set -e

echo "Building Go app"
go build -ldflags="-X 'main.twitchClientID=${TWITCH_CLIENT_ID}' -X 'main.twitchClientSecret=${TWITCH_SECRET_ID}'" -tags production -o out/twitch-clip .

echo "Packaging macOS app"
go run scripts/macapp.go -assets ./out -bin twitch-clip -icon ./assets/icon1080.png -identifier com.skynewz.twitchclip -name "Twitch Clip" &&
  mv "Twitch Clip.app" out
