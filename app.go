package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/SkYNewZ/twitch-clip/internal/icon"
	"github.com/SkYNewZ/twitch-clip/internal/twitch"
	"github.com/SkYNewZ/twitch-clip/pkg/player"
	"github.com/SkYNewZ/twitch-clip/pkg/streamlink"
	"github.com/atotto/clipboard"
	"github.com/emersion/go-autostart"
	"github.com/getlantern/systray"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
)

var (
	// These 2 variables will be overwritten at build time using ldflags
	twitchClientID     string
	twitchClientSecret string
)

const (
	AppName        = "twitchclip"
	AppDisplayName = "Twitch Clip"
)

// application contains all required dependencies
type application struct {
	Name        string
	DisplayName string

	// Main cancel function to stop the program
	Cancel context.CancelFunc

	// Player to use
	Player player.Player

	// Twitch client
	Twitch *twitch.Client

	Streamlink streamlink.Client

	// Carry our current displayed items
	State map[string]*Item

	// Each string in this chan will be send to system clipboard
	ClipboardListener chan string
}

// New creates a new application
func New() *application {
	// Get media player
	p, err := player.DefaultPlayer()
	if err != nil {
		log.Fatalln(err)
	}

	s, err := streamlink.New()
	if err != nil {
		log.Fatalln(err)
	}

	// Get Twitch client
	twitchClient, err := twitch.New(&twitch.Config{ClientID: twitchClientID, ClientSecret: twitchClientSecret})
	if err != nil {
		log.Fatalln(err)
	}

	// Make the app and inject required dependencies
	return &application{
		Name:              AppName,
		DisplayName:       AppDisplayName,
		Cancel:            nil,
		Player:            p,
		Twitch:            twitchClient,
		State:             make(map[string]*Item),
		ClipboardListener: make(chan string, 1),
		Streamlink:        s,
	}
}

// Setup must not be called before systray.Run or systray.Register
// app := New()
// systray.Run(app.Setup, app.Stop)
func (a *application) Setup() {
	// Set icon and application name
	systray.SetIcon(icon.Data)
	systray.SetTooltip(a.DisplayName)

	// Manage auto start
	done := make(chan struct{}, 1)
	go a.autostart(done)
	<-done //wait for "autostart" button displayed before continue

	// Display "quit" button and listen for click
	quit := systray.AddMenuItem("Quit", "Quit the whole app")
	systray.AddSeparator()
	go func() {
		<-quit.ClickedCh
		systray.Quit()
	}()

	// Start application
	a.Start()
}

// Start show a Item for each online streams
// This will be refresh at each streamsRefreshTime
// The passed context is used to cancel theses routines
func (a *application) Start() {
	var ctx context.Context
	ctx, a.Cancel = context.WithCancel(context.Background())

	// We permit only one array at a time
	var out = make(chan []*twitch.Stream, 1)

	// start routines for refreshing streams
	go a.RefreshActiveStreams(ctx, out)

	// start routine to display these streams
	go a.RefreshStreamsMenuItem(ctx, out)

	// Listen for clipboard requests
	go a.HandleClipboard(ctx)
}

// Stop application
func (a *application) Stop() {
	a.Cancel()
	close(a.ClipboardListener)
}

// autostart make current application auto start at boot and handle change on the item
func (a *application) autostart(done chan<- struct{}) {
	executable, err := os.Executable()
	if err != nil {
		log.Warningf("cannot find current executable file path. Application won't start automatically: %s", err)
		return
	}

	app := &autostart.App{
		Name:        a.Name,
		DisplayName: a.DisplayName,
		Exec:        []string{executable},
	}

	autostartItem := systray.AddMenuItemCheckbox("Start at login", "Start this app at system startup", app.IsEnabled())
	close(done) // autostartItem is displayed, we have done

	for {
		<-autostartItem.ClickedCh // wait for a click
		switch app.IsEnabled() {
		case true:
			log.Debugln("disable application autostart")
			if err := app.Disable(); err != nil {
				log.Errorln(err)
				continue
			}

			autostartItem.Uncheck()
		case false:
			log.Debugln("enable application autostart")
			if err := app.Enable(); err != nil {
				log.Errorln(err)
				continue
			}

			autostartItem.Check()
		}
	}
}

// ClickStartOnStartup manage if this app must start on system startup or not
func (a *application) ClickStartOnStartup(item *systray.MenuItem, app *autostart.App) {
	for {
		<-item.ClickedCh // wait for a click
		switch app.IsEnabled() {
		case true:
			log.Println("disable application autostart")
			if err := app.Disable(); err != nil {
				log.Errorln(err)
				continue
			}

			item.Uncheck()
		case false:
			log.Println("enable application autostart")
			if err := app.Enable(); err != nil {
				log.Errorln(err)
				continue
			}

			item.Check()
		}
	}
}

func (a *application) HandleClipboard(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debugln("received context cancel: HandleClipboard")
			return // returning not to leak the goroutine
		case link := <-a.ClipboardListener:
			log.Tracef("setting [%s] to clipboard", link)
			if err := clipboard.WriteAll(link); err != nil {
				log.Errorln(err)
				continue // do not stop this routine in case of error
			}
		}
	}
}

// Refresh hide or show menu items based on currently active streams
func (a *application) Refresh(activeStreams []*twitch.Stream) {
	var wg sync.WaitGroup
	wg.Add(len(a.State))
	for _, item := range a.State {
		go func(i *Item) {
			defer wg.Done()
			itemIsAnActiveStream := funk.Contains(activeStreams, func(stream *twitch.Stream) bool {
				return stream.UserLogin == i.ID
			})

			i.SetVisible(itemIsAnActiveStream)
		}(item)
	}

	wg.Wait()
}

func (a *application) DisplayConnectedUser() {
	me := a.Twitch.Users.Me()
	title := fmt.Sprintf("Connected as %s", me.DisplayName)
	systray.AddMenuItem(title, "Current user").Disable()
}

// RefreshActiveStreams send active streams to out
func (a *application) RefreshActiveStreams(ctx context.Context, out chan<- []*twitch.Stream) {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	job := func() {
		log.Debugln("refreshing followed streams infos")

		// This simulates /streams/followed endpoint
		streams, err := a.Twitch.Streams.GetFollowed()
		if err != nil {
			log.Errorf("unable to list followed streams: %s", err)
			return
		}

		// job done, notify out for the new stream list
		out <- streams
	}

	// https://stackoverflow.com/a/54752803
	for {
		job()

		select {
		case <-ctx.Done():
			log.Debugln("received context cancel: RefreshActiveStreams")
			return // returning not to leak the goroutine
		case <-ticker.C:
			continue
		}
	}
}

// RefreshStreamsMenuItem display a menu Item for each stream received in the channel in
func (a *application) RefreshStreamsMenuItem(ctx context.Context, in <-chan []*twitch.Stream) {
	// not active stream menu Item
	menuNoActiveStreams := &Item{
		Application: nil,
		Item:        systray.AddMenuItem("No active stream", "No active stream"),
		Visible:     true,
		ID:          "",
		mutex:       sync.Mutex{},
	}
	menuNoActiveStreams.Disable()

	// display connected user
	a.DisplayConnectedUser()

	for {
		select {
		case <-ctx.Done():
			log.Debugln("received context cancel: RefreshStreamsMenuItem")
			return // returning not to leak the goroutine
		case activeStreams := <-in:
			log.Debugf("refreshing menu items for %d active followed streams", len(activeStreams))
			menuNoActiveStreams.SetVisible(len(activeStreams) == 0)

			for _, s := range activeStreams {
				tooltip := s.Title
				title := fmt.Sprintf("%s (%s)", s.UserName, s.GameName)

				// stream already in the stream list. Refresh title and tooltip and show it
				if v, ok := a.State[s.UserLogin]; ok {
					v.Refresh(title, tooltip)
					continue
				}

				// stream not already in the stream list, make it!
				a.State[s.UserLogin] = a.NewItem(ctx, s, title, tooltip)
			}

			// refresh app
			a.Refresh(activeStreams)
		}
	}
}

// NewItem creates a new menu Item and its routine click
func (a *application) NewItem(ctx context.Context, s *twitch.Stream, title, tooltip string) *Item {
	log.WithFields(map[string]interface{}{
		"login":    s.UserLogin,
		"user_id":  s.ID,
		"username": s.UserName,
		"game":     s.GameName,
	}).Tracef("new active stream detected [%s]", s.UserLogin)

	item := &Item{
		Item:        systray.AddMenuItem(title, tooltip),
		Visible:     true, // Visible by default
		ID:          s.UserLogin,
		mutex:       sync.Mutex{},
		Application: a,
	}

	// Start routine to pull its icon
	go item.SetIcon()

	// Start routine click for this Item
	go item.Click(ctx)

	return item
}
