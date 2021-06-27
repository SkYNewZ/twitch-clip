package streamlink

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var _ Client = (*client)(nil)

var (
	once       sync.Once
	streamlink *client
)

type Client interface {
	// Run gets the given streamer's stream URL
	Run(streamer string) ([]byte, error)
}

type client struct {
	Options []string
}

// New create a Client instance
func New(opts ...string) (Client, error) {
	var err error
	once.Do(func() {
		streamlink = new(client)
		streamlink.Options = append([]string{
			"streamlink",           // streamlink binary
			"--quiet",              // Hide all log output.
			"--twitch-low-latency", // enable Twitch low latency for supported stream https://streamlink.github.io/cli.html#cmdoption-twitch-low-latency
			"--stream-url",         // https://streamlink.github.io/cli.html#cmdoption-stream-url
			"--twitch-disable-ads", // disable Twitch ads https://streamlink.github.io/cli.html#cmdoption-twitch-disable-ads
			"$url",                 // Place holder to place URL here
			"best",                 // best video source
		}, opts...)

		// Search in path
		if _, err = exec.LookPath(streamlink.Options[0]); err != nil {
			return
		}
	})

	return streamlink, nil
}

func (c *client) Run(streamer string) ([]byte, error) {
	// Fill arguments with URL
	var tmpCommand = make([]string, len(c.Options))
	copy(tmpCommand, c.Options)
	for i := range tmpCommand {
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$url", fmt.Sprintf("https://www.twitch.tv/%s", streamer))

	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// run cmd with a timeout of 10 seconds
	cmd := exec.CommandContext(ctx, streamlink.Options[0], tmpCommand[1:]...)
	log.Debugf("running command [%s]", cmd.String())
	return cmd.Output()
}
