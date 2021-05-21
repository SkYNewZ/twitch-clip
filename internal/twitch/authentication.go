package twitch

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2/twitch"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

var _ http.RoundTripper = (*transport)(nil)

const (
	tokenDir  = "Twitch Clip"
	tokenFile = "token.dat" // used to store current session token on disk
)

type transport struct {
	Original http.RoundTripper
	clientID string // must be in each request
}

func (t transport) RoundTrip(r *http.Request) (*http.Response, error) {
	log.Debugf("URL being requested: %s", r.URL.String())
	r.Header.Set("Client-Id", t.clientID)
	return t.Original.RoundTrip(r)
}

var (
	state              string
	srv                *http.Server                 // server to handle redirect URI
	authenticationDone = make(chan *http.Client, 1) // notify when callback process is done, we receive the configured http.Client
	oauth2Config       *oauth2.Config               // carry the entire Twitch oauth2 process

	// ErrInvalidState state configured between request and response
	ErrInvalidState = errors.New("invalid state coming from Twitch")
)

// handleUserLogin creates the initial URL request
// User's browser will be open with this URL
func handleUserLogin(oauth2Config *oauth2.Config) error {
	var tokenBytes [255]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return fmt.Errorf("unable to generate random state: %w", err)
	}

	// save state as global to protect again CSRF
	state = hex.EncodeToString(tokenBytes[:])
	u := oauth2Config.AuthCodeURL(state)
	return browser.OpenURL(u)
}

// handleOAuth2Callback is a Handler for oauth's 'redirect_uri' endpoint;
// it validates the state token and retrieves an OAuth token from the request parameters.
func handleOAuth2Callback(w http.ResponseWriter, r *http.Request) error {
	log.Debugln("received oauth2 callback")
	if v := r.FormValue("state"); v != state {
		return ErrInvalidState
	}

	// Use the custom HTTP client when requesting a token.
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, setupHTTPClient(oauth2Config.ClientID))
	token, err := oauth2Config.Exchange(ctx, r.FormValue("code"))
	if err != nil {
		return fmt.Errorf("unable to oauth code: %w", err)
	}

	// send our configured http.Client
	authenticationDone <- oauth2Config.Client(ctx, token)

	// Store token on disk
	go func() {
		if err := storeTokenOnFile(token); err != nil {
			log.Errorln(err)
		}
	}()

	// Close this web server, we don't need it anymore
	go func() {
		time.Sleep(time.Second * 10) // let user see the response
		log.Debugf("closing web server")
		if err := srv.Close(); err != nil {
			log.Errorln(err)
		}
	}()

	_, _ = fmt.Fprint(w, "Authentication successful, you can close this tab")
	return nil
}

// configOAuth2Workflow create a oauth2.Config object related to Twitch
// It will create and start an http.Server to handle redirect URI
// By default, it will use a free open port for the redirect URI
// The returned http.Server must be Close() when job is done
// https://github.com/twitchdev/authentication-go-sample/blob/main/oauth-authorization-code/main.go
func configOAuth2Workflow() {
	var errorHandling = func(handler func(http.ResponseWriter, *http.Request) error) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := handler(w, r); err != nil {
				log.Errorln(err)
				var errorString = "Something went wrong! Please try again."
				http.Error(w, errorString, http.StatusInternalServerError)
				return
			}
		})
	}

	addr := "localhost:7001"
	mux := http.NewServeMux()
	mux.Handle("/", errorHandling(handleOAuth2Callback))
	srv = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// start server
	go func() {
		log.Debugf("starting web server for oauth2 callblack at %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("fail to stop web server: %s", err)
		}
	}()
}

func getToken(ctx context.Context, config *Config) (*http.Client, error) {
	oauth2Config = &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Scopes:       []string{"user:read:follows"},
		Endpoint:     twitch.Endpoint,
		RedirectURL:  "http://localhost:7001",
	}

	// Retrieve token from disk
	token, err := retrieveTokenOnFile()
	if err == nil {
		// We have our token on disk, use it!
		log.Debugln("using token from disk")
		c := context.WithValue(context.Background(), oauth2.HTTPClient, setupHTTPClient(config.ClientID))
		return oauth2Config.Client(c, token), nil
	}

	// No token found, we need a new one
	log.Warningln(err)

	// setup server
	configOAuth2Workflow()

	// Start the oauth workflow
	if err := handleUserLogin(oauth2Config); err != nil {
		return nil, err
	}

	// Wait for authentication process
	log.Debugln("waiting for authentication callback")
	select {
	case v := <-authenticationDone:
		return v, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// setupHTTPClient return our custom HTTP client
func setupHTTPClient(clientID string) *http.Client {
	return &http.Client{
		Timeout:   time.Second * 10, // global timeout
		Transport: &transport{http.DefaultTransport, clientID},
	}
}

// storeTokenOnFile save given token to file
func storeTokenOnFile(token *oauth2.Token) error {
	configDir, _ := os.UserConfigDir()
	configDir = filepath.Join(configDir, tokenDir)

	// crete config directory
	if _, err := os.Stat(configDir); errors.Is(err, os.ErrNotExist) {
		log.Tracef("creating %s", configDir)
		_ = os.Mkdir(configDir, 0755)
	}

	tokenFilePath := filepath.Join(configDir, tokenFile)
	// If the file doesn't exist, create it, or append to the file
	log.Tracef("opening %s", tokenFilePath)
	f, err := os.OpenFile(tokenFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("unable to open token file: %w", err)
	}
	defer f.Close()

	log.Debugf("writing token content to %s", tokenFilePath)
	return json.NewEncoder(f).Encode(token)
}

// retrieveTokenOnFile return token store on disk
func retrieveTokenOnFile() (*oauth2.Token, error) {
	configDir, _ := os.UserConfigDir()
	configDir = filepath.Join(configDir, tokenDir)

	// If folder does not exist, return error
	if _, err := os.Stat(configDir); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("cannot find token file in %s", configDir)
	}

	tokenFilePath := filepath.Join(configDir, tokenFile)
	log.Tracef("opening %s", tokenFilePath)
	f, err := os.OpenFile(tokenFilePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("cannot open token file: %w", err)
	}

	token := new(oauth2.Token)
	log.Tracef("reading %s", tokenFilePath)
	if err := json.NewDecoder(f).Decode(token); err != nil {
		return nil, fmt.Errorf("cannot read token file: %w", err)
	}

	return token, nil
}
