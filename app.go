package main

//go:generate go-winres make --in assets/winres.json --product-version ${VERSION}
//go:generate sh -c "INPUT=assets/icon22.png OUTPUT=internal/icon/icon_unix.go scripts/make_icon.sh"
//go:generate sh -c "INPUT=assets/icon.ico OUTPUT=internal/icon/icon_windows.go scripts/make_icon.sh"

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/SkYNewZ/twitch-clip/internal/icon"
	"github.com/SkYNewZ/twitch-clip/internal/twitch"
	"github.com/SkYNewZ/twitch-clip/pkg/notifier"
	"github.com/SkYNewZ/twitch-clip/pkg/player"
	"github.com/SkYNewZ/twitch-clip/pkg/streamlink"
	"github.com/atotto/clipboard"
	"github.com/emersion/go-autostart"
	"github.com/getlantern/systray"
	"github.com/pkg/browser"
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

// Application contains all required dependencies
type Application struct {
	Name        string
	DisplayName string

	// Main cancel function to stop the program
	Cancel context.CancelFunc

	// Player to use
	Player player.Player

	// Twitch client
	Twitch *twitch.Client

	Streamlink streamlink.Client

	Notifier               notifier.Notifier
	NotificationCallbackCh <-chan string

	// Carry our current displayed items
	State map[string]*Item

	// Each string in this chan will be send to system clipboard
	ClipboardListener chan string
}

// New creates a new Application
func New() *Application {
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

	// Start the notifier service
	n, notificationCh := notifier.New(AppDisplayName)

	// Make the app and inject required dependencies
	return &Application{
		Name:                   AppName,
		DisplayName:            AppDisplayName,
		Cancel:                 nil,
		Player:                 p,
		Twitch:                 twitchClient,
		Streamlink:             s,
		Notifier:               n,
		NotificationCallbackCh: notificationCh,
		State:                  make(map[string]*Item),
		ClipboardListener:      make(chan string, 1),
	}
}

// Setup must not be called before systray.Run or systray.Register
// app := New()
// systray.Run(app.Setup, app.Stop)
func (a *Application) Setup() {
	// Set icon and Application name
	systray.SetIcon(icon.Data)
	systray.SetTooltip(a.DisplayName)

	// Manage auto start
	done := make(chan struct{}, 1)
	go a.autostart(done)
	<-done //wait for "autostart" button displayed before continue

	// This context will manage all Application routines cancellation
	var ctx context.Context
	ctx, a.Cancel = context.WithCancel(context.Background())

	// Open Twitch website
	openTwitch := systray.AddMenuItem("Open Twitch", "Open https://www.twitch.tv")
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-openTwitch.ClickedCh:
				if err := browser.OpenURL("https://www.twitch.tv"); err != nil {
					log.Errorf("unable to open twitch website: %s", err)
				}
			}
		}
	}()

	// Display "quit" button and listen for click
	quit := systray.AddMenuItem("Quit", "Quit the whole app")
	systray.AddSeparator()
	go func() {
		<-quit.ClickedCh
		systray.Quit()
	}()

	// Start Application
	a.Start(ctx)
}

// Start show a Item for each online streams
// This will be refresh at each streamsRefreshTime
// The passed context is used to cancel theses routines
func (a *Application) Start(ctx context.Context) {
	// We permit only one array at a time
	var out = make(chan []*twitch.Stream, 1)

	// Listen for notification callback
	go a.HandleNotificationCallback(ctx)

	// start routines for refreshing streams
	go a.RefreshActiveStreams(ctx, out)

	// start routine to display these streams
	go a.RefreshStreamsMenuItem(ctx, out)

	// Listen for clipboard requests
	go a.HandleClipboard(ctx)
}

// Stop Application
func (a *Application) Stop() {
	a.Cancel()                                 // stop each routines
	close(a.ClipboardListener)                 // stop clipboard listener
	if err := a.Notifier.Close(); err != nil { // notification service
		log.Errorf("fail to stop notification service: %s", err)
	}
}

// autostart make current Application auto start at boot and handle change on the item
func (a *Application) autostart(done chan<- struct{}) {
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
			log.Debugln("disable Application autostart")
			if err := app.Disable(); err != nil {
				log.Errorln(err)
				continue
			}

			autostartItem.Uncheck()
		case false:
			log.Debugln("enable Application autostart")
			if err := app.Enable(); err != nil {
				log.Errorln(err)
				continue
			}

			autostartItem.Check()
		}
	}
}

func (a *Application) HandleClipboard(ctx context.Context) {
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
func (a *Application) Refresh(activeStreams []*twitch.Stream) {
	var wg sync.WaitGroup
	wg.Add(len(a.State))
	for _, item := range a.State {
		go func(i *Item) {
			defer wg.Done()
			itemIsAnActiveStream := funk.Contains(activeStreams, func(stream *twitch.Stream) bool {
				return stream.UserLogin == i.UserLogin
			})

			i.SetVisible(itemIsAnActiveStream)
		}(item)
	}

	wg.Wait()
}

func (a *Application) DisplayConnectedUser() {
	me := a.Twitch.Users.Me()
	title := fmt.Sprintf("Connected as %s", me.DisplayName)
	systray.AddMenuItem(title, "Current user").Disable()
}

// RefreshActiveStreams send active streams to out
func (a *Application) RefreshActiveStreams(ctx context.Context, out chan<- []*twitch.Stream) {
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
func (a *Application) RefreshStreamsMenuItem(ctx context.Context, in <-chan []*twitch.Stream) {
	// not active stream menu Item
	menuNoActiveStreams := &Item{
		Application: nil,
		Item:        systray.AddMenuItem("No active stream", "No active stream"),
		Visible:     true,
		UserLogin:   "",
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
			if len(activeStreams) == 0 {
				continue // no active stream, just leave here
			}

			for _, s := range activeStreams {
				// stream already in the stream list. Refresh title and tooltip and show it
				if v, ok := a.State[s.UserLogin]; ok {
					v.Refresh(s)
					continue
				}

				// stream not already in the stream list, make it!
				a.State[s.UserLogin] = a.NewItem(ctx, s)
			}

			// refresh app
			a.Refresh(activeStreams)
		}
	}
}

// NewItem creates a new menu Item and its underlying routines
func (a *Application) NewItem(ctx context.Context, s *twitch.Stream) *Item {
	log.WithFields(map[string]interface{}{
		"login":      s.UserLogin,
		"user_login": s.UserName,
		"username":   s.UserName,
		"game":       s.GameName,
	}).Tracef("new active stream detected [%s]", s.UserLogin)

	// sometimes the Twitch API does not send the username at first call, use the user UserLogin instead
	username := s.UserName
	if username == "" {
		username = s.UserLogin
	}

	item := &Item{
		Application: a,
		Item:        systray.AddMenuItem(fmt.Sprintf("%s (%s)", username, s.GameName), s.Title),
		Visible:     true, // Visible by default
		UserLogin:   s.UserLogin,
		Username:    s.UserName,
		Game:        s.GameName,
		mutex:       sync.Mutex{},
	}

	// Start routine to pull its icon
	go item.SetIcon()

	// Start routine click for this Item
	go item.Click(ctx)

	// New item appear, so notify
	if err := a.Notifier.Notify(username, item.Game, item.UserLogin); err != nil {
		log.Errorf("fail to notify for [%s]: %s", item.UserLogin, err)
	}

	return item
}

// HandleNotificationCallback receives notification callback and launch streamlink process
func (a *Application) HandleNotificationCallback(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debugln("received context cancel: HandleNotificationCallback")
			return // returning not to leak the goroutine
		case v := <-a.NotificationCallbackCh:
			// get menu item matching streamer name
			item, ok := a.State[v]
			if !ok {
				log.Errorf("received notification callback for non-existent stream [%s]", v)
				continue
			}

			// simulate a click
			item.Item.ClickedCh <- struct{}{}
		}
	}
}
