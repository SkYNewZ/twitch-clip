package twitch

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/peterbourgon/diskv/v3"
)

const (
	apiURL   = "https://api.twitch.tv/helix"
	cacheDir = "Twitch Clip"
)

type Client struct {
	httpClient *http.Client // make each Twitch requests. Requests will be authenticated
	Me         *User        // current connected user
	cache      *diskv.Diskv // store avatar

	// Available public methods on client
	Streams StreamsI
	Users   UsersI
}

type Config struct {
	ClientID     string
	ClientSecret string
}

// New returns a new Twitch client
func New(config *Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("missing Twitch config config")
	}

	if config.ClientID == "" {
		return nil, fmt.Errorf("missing Twitch client ID. Check https://dev.twitch.tv/console/apps/create")
	}

	if config.ClientSecret == "" {
		return nil, fmt.Errorf("missing Twitch client secret. Check https://dev.twitch.tv/console/apps/create")
	}

	// create the client
	c := new(Client)

	// create cache
	var err error
	c.cache, err = createCacheDir()
	if err != nil {
		return nil, fmt.Errorf("unable to create cache directory: %w", err)
	}

	// Wait for http.Client
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	c.httpClient, err = getToken(ctx, config)
	if err != nil {
		return nil, err
	}

	// Get the original transport
	c.Streams = &streamsClient{c}
	c.Users = &usersClient{c}

	// Search for current user info
	u, err := c.Users.Get()
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Twitch client: %w", err)
	}

	c.Me = u[0]
	return c, nil
}

func createCacheDir() (*diskv.Diskv, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("unable to find cache directory: %w", err)
	}

	return diskv.New(diskv.Options{
		BasePath:     filepath.Join(dir, cacheDir),
		CacheSizeMax: 20 * 10 * 1024, // 20 images, each image is 10 KB
		PathPerm:     0755,
		FilePerm:     0644,
	}), nil
}
