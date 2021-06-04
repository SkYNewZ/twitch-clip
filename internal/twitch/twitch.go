package twitch

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/peterbourgon/diskv/v3"
)

const (
	apiURL   = "https://api.twitch.tv/helix"
	cacheDir = "Twitch Clip"
)

type Client struct {
	httpClient *http.Client // make each Twitch requests. Requests will be authenticated
	cache      *diskv.Diskv // store avatar
	me         *User        // current connected user

	// Available public methods on client
	Streams StreamsI
	Users   UsersI
}

type Config struct {
	ClientID     string
	ClientSecret string
}

var (
	// Singleton
	once   sync.Once
	client *Client
)

// New returns a new Twitch client
func New(config *Config) (*Client, error) {
	var err error
	once.Do(func() {
		if config == nil {
			err = fmt.Errorf("missing Twitch config config")
			return
		}

		if config.ClientID == "" {
			err = fmt.Errorf("missing Twitch client ID. Check https://dev.twitch.tv/console/apps/create")
			return
		}

		if config.ClientSecret == "" {
			err = fmt.Errorf("missing Twitch client secret. Check https://dev.twitch.tv/console/apps/create")
			return
		}

		// create the client
		client = new(Client)

		// create cache
		client.cache, err = createCacheDir()
		if err != nil {
			err = fmt.Errorf("unable to create cache directory: %w", err)
			return
		}

		// Wait for http.Client
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		client.httpClient, err = getToken(ctx, config)
		if err != nil {
			return
		}

		// Get the original transport
		client.Streams = &streamsClient{client}
		client.Users = &usersClient{client}

		// Get current connected user
		var users []*User
		users, err = client.Users.Get()
		if err != nil {
			err = fmt.Errorf("unable to initialize client: %w", err)
			return
		}

		client.me = users[0]
	})

	return client, err
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
