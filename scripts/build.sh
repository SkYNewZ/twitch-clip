#!/usr/bin/env bash

echo "Cleaning"
rm -rf assets/twitch-clip Twitch\ Clip.app

echo "Build Go app"
mkdir -p assets
go build -ldflags="-X 'main.twitchClientID=${TWITCH_CLIENT_ID}' -X 'main.twitchClientSecret=${TWITCH_SECRET_ID}'" -tags production -o assets/twitch-clip .

echo "Package app"
pwd
go run scripts/macapp.go -assets ./assets -bin twitch-clip -icon ./internal/icon/logo.png -identifier com.skynewz.twitchclip -name "Twitch Clip"
