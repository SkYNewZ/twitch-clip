#!/usr/bin/env bash

echo "Cleaning"
rm -rf assets/twitch-clip Twitch\ Clip.app

echo "Build Go app"
mkdir assets
go build -tags production -o assets/twitch-clip .

echo "Package app"
go run scripts/macapp.go -assets ./assets -bin twitch-clip -icon ./internal/icon/logo.png -identifier com.skynewz.twitchclip -name "Twitch Clip"
