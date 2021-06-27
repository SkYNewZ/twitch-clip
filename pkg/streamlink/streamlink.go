package streamlink

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	_ Client = (*client)(nil)

	// ErrStreamLinkNotFound .
	ErrStreamLinkNotFound = errors.New("streamlink not found in PATH. Check https://streamlink.github.io/install.html")
)

type Client interface {
	// Run gets the given streamer's stream URL
	Run(streamer string) ([]byte, error)
}

// client implements Client interface
type client struct {
	Options []string
}

// New create a Client instance
func New(opts ...string) (Client, error) {
	streamlink := new(client)
	streamlink.Options = append([]string{
		"streamlink",           // streamlink binary
		"--quiet",              // Hide all log output.
		"--twitch-low-latency", // enable Twitch low latency for supported stream https://streamlink.github.io/cli.html#cmdoption-twitch-low-latency
		"--stream-url",         // https://streamlink.github.io/cli.html#cmdoption-stream-url
		"--twitch-disable-ads", // disable Twitch ads https://streamlink.github.io/cli.html#cmdoption-twitch-disable-ads
		"$url",                 // Place holder to place URL here
		"best",                 // best video source
	}, opts...) // append user-defined options

	// Search in path
	v, err := exec.LookPath(streamlink.Options[0])
	if err != nil {
		return nil, ErrStreamLinkNotFound
	}

	// Replace with absolute command path
	log.Tracef("found [streamlink] at [%s]", v)
	streamlink.Options[0] = v
	return streamlink, nil
}

func (c *client) Run(streamer string) ([]byte, error) {
	// Fill arguments with URL
	var tmpCommand = make([]string, len(c.Options))
	copy(tmpCommand, c.Options)
	for i := range tmpCommand {
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$url", fmt.Sprintf("https://www.twitch.tv/%s", streamer))

	}

	// run cmd with a timeout of 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.Options[0], tmpCommand[1:]...)
	log.Debugf("running command [%s]", cmd.String())
	return cmd.Output()
}
