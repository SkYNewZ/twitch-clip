package streamlink

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

const streamlinkBin = "/usr/local/bin/streamlink"

var streamLinkArgs = []string{
	"best",                 // best video source
	"--twitch-low-latency", // enable Twitch low latency for supported stream https://streamlink.github.io/cli.html#cmdoption-twitch-low-latency
	"--stream-url",         // https://streamlink.github.io/cli.html#cmdoption-stream-url
	"--twitch-disable-ads", // disable Twitch ads https://streamlink.github.io/cli.html#cmdoption-twitch-disable-ads
}

// InPath check whether streamlink is in path
func InPath() bool {
	_, err := exec.LookPath(streamlinkBin)
	return err == nil
}

// ExecuteStreamlinkCommand gets the given streamer's stream URL
func ExecuteStreamlinkCommand(streamer string) ([]byte, error) {
	twitchURL := fmt.Sprintf("https://www.twitch.tv/%s", streamer)
	args := append([]string{twitchURL}, streamLinkArgs...)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// run cmd with a timeout oof 10 seconds
	cmd := exec.CommandContext(ctx, streamlinkBin, args...)
	log.Debugf("running command [%s]", cmd.String())
	return cmd.Output()
}
