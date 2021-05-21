package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SkYNewZ/twitch-clip/internal/iina"
	"github.com/SkYNewZ/twitch-clip/internal/twitch"
	"github.com/atotto/clipboard"
	"github.com/getlantern/systray"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
)

var (
	twitchClient       *twitch.Client
	streamsRefreshTime = time.Second * 10
)

type menuItem struct {
	MenuItem *systray.MenuItem
	Visible  bool
	ID       string
	m        sync.Mutex
}

// Show item if not already visible
func (i *menuItem) Show() {
	i.m.Lock()
	defer i.m.Unlock()
	if i.Visible {
		return
	}

	i.MenuItem.Show()
	i.Visible = true
}

// Hide item if not already hidden
func (i *menuItem) Hide() {
	i.m.Lock()
	defer i.m.Unlock()
	if !i.Visible {
		return
	}

	i.MenuItem.Hide()
	i.Visible = false
}

// SetVisible set whether current item should be visible
func (i *menuItem) SetVisible(visible bool) {
	switch visible {
	case true:
		i.Show()
	case false:
		i.Hide()
	}
}

// Refresh refresh item info
func (i *menuItem) Refresh(title, tooltip string) {
	i.MenuItem.SetTitle(title)
	i.MenuItem.SetTooltip(tooltip)
}

func (i *menuItem) Disable() {
	i.MenuItem.Disable()
}

func (i *menuItem) handleStreamMenuItemClick(ctx context.Context) {
	log.Debugf("starting click routine for [%s]", i.ID)
	for {
		select {
		case <-ctx.Done():
			log.Debugf("received context cancel: handleStreamMenuItemClick [%s]", i.ID)
			return // returning not to leak the goroutine
		case <-i.MenuItem.ClickedCh:
			log.Debugf("[%s] item is clicked", i.ID)

			// Get link
			u, err := getStreamLink(i.ID)
			if err != nil {
				log.Errorln(err)
				return
			}

			// Setting in clipboard
			log.Debugf("setting link to clipboard for [%s]", i.ID)
			if err := clipboard.WriteAll(u); err != nil {
				log.Errorln(err)
				return
			}

			// open with iina
			log.Debugf("openning with iina for [%s]", i.ID)
			if err := iina.Run(u); err != nil {
				log.Errorln(err)
				return
			}
		}
	}
}

// setUserIcon pull avatar and set to given menu item
func (i *menuItem) setUserIcon() {
	users, err := twitchClient.Users.Get(i.ID)
	if err != nil {
		log.Errorf("unable to refresh Twitch user info for %s: %s", i.ID, err)
		return
	}

	if len(users) == 0 {
		log.Warningf("no image found for %s", i.ID)
		return
	}

	// get user icon
	img, err := twitchClient.Users.ProfileImageBytes(users[0])
	if err != nil {
		log.Errorln(err)
		return
	}

	// set icon
	i.MenuItem.SetIcon(img)
}

// RefreshVisible hide or show current menu items based on currentLiveStreams
func (i *menuItem) RefreshVisible(activeStreams []*twitch.Stream) {
	itemIsAnActiveStream := funk.Contains(activeStreams, func(stream *twitch.Stream) bool {
		return stream.UserLogin == i.ID
	})

	i.SetVisible(itemIsAnActiveStream)
}

func setupTwitch() {
	var err error
	twitchClient, err = twitch.New(&twitch.Config{ClientID: "[REDACTED]", ClientSecret: "[REDACTED]"})
	if err != nil {
		log.Fatalln(err)
	}
}

func displayConnectedUser() {
	title := fmt.Sprintf("Connected as %s", twitchClient.Me.DisplayName)
	systray.AddMenuItem(title, "Current user").Disable()
}

// setupStreamsMenuItem show a item for each online streams
// This will be refresh at each streamsRefreshTime
// The passed context is used to cancel theses routines
func setupStreamsMenuItem(ctx context.Context) {
	var out = make(chan []*twitch.Stream)

	// start routines for refreshing streams
	go refreshActiveStreams(ctx, out)

	// start routine to display these streams
	go refreshStreamsMenuItem(ctx, out)
}

// refreshActiveStreams send active streams to out
func refreshActiveStreams(ctx context.Context, out chan<- []*twitch.Stream) {
	ticker := time.NewTicker(streamsRefreshTime)
	defer ticker.Stop()

	job := func() {
		log.Debugln("refreshing followed streams infos")

		// This simulates /streams/followed endpoint
		streams, err := twitchClient.Streams.GetFollowed()
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
			log.Debugln("received context cancel: refreshActiveStreams")
			return // returning not to leak the goroutine
		case <-ticker.C:
			continue
		}
	}
}

// refreshStreamsMenuItem display a menu item for each stream received in the channel in
func refreshStreamsMenuItem(ctx context.Context, in <-chan []*twitch.Stream) {
	// Used to manage menu items
	var streamsCh = map[string]*menuItem{}

	// not active stream menu item
	menuNoActiveStreams := &menuItem{
		MenuItem: systray.AddMenuItem("No active stream", "No active stream"),
		Visible:  true,
	}
	menuNoActiveStreams.Disable()

	// display connected user
	displayConnectedUser()

	for {
		select {
		case <-ctx.Done():
			log.Debugln("received context cancel: refreshStreamsMenuItem")
			return // returning not to leak the goroutine
		case activeStreams := <-in:
			log.Debugf("refreshing menu items for %d active followed streams", len(activeStreams))
			menuNoActiveStreams.SetVisible(len(activeStreams) == 0)

			for _, s := range activeStreams {
				tooltip := s.Title
				title := fmt.Sprintf("%s (%s)", s.UserName, s.GameName)

				// stream already in the stream list. Refresh title and tooltip and show it
				if v, ok := streamsCh[s.UserLogin]; ok {
					v.Refresh(title, tooltip)
					continue
				}

				// stream not already in the stream list, make it!
				streamsCh[s.UserLogin] = makeNewMenuItem(ctx, s, title, tooltip)
			}

			// show/hide menu item
			for _, item := range streamsCh {
				go item.RefreshVisible(activeStreams)
			}
		}
	}
}

// makeNewMenuItem creates a new menu item and its routine click
func makeNewMenuItem(ctx context.Context, s *twitch.Stream, title, tooltip string) (item *menuItem) {
	log.WithFields(map[string]interface{}{
		"login":    s.UserLogin,
		"user_id":  s.ID,
		"username": s.UserName,
		"game":     s.GameName,
	}).Tracef("new active stream detected [%s]", s.UserLogin)

	item = &menuItem{
		MenuItem: systray.AddMenuItem(title, tooltip),
		ID:       s.UserLogin,
		Visible:  true, // visible by default
	}

	// Start routine to pull its icon
	go item.setUserIcon()

	// Start routine click for this item
	go item.handleStreamMenuItemClick(ctx)
	return
}
