package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/SkYNewZ/twitch-clip/internal/twitch"

	"github.com/getlantern/systray"
	log "github.com/sirupsen/logrus"
)

// Item describes a displayed item
type Item struct {
	Application *Application
	Item        *systray.MenuItem
	Visible     bool
	UserLogin   string // streamer user UserLogin (e.g. locklear)
	Username    string // streamer displayed username (e.g. Locklear)
	Game        string // game name on stream (e.g. Just Chatting)
	mutex       sync.Mutex
}

// Show Item if not already Visible
func (i *Item) Show() {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if i.Visible {
		return
	}

	i.Item.Show()
	i.Visible = true

	// Item becomes visible, notify it
	if err := i.Application.Notifier.Notify(i.Username, i.Game, i.UserLogin); err != nil {
		log.Errorf("fail to notify for [%s]: %s", i.UserLogin, err)
	}
}

// Hide Item if not already hidden
func (i *Item) Hide() {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if !i.Visible {
		return
	}

	i.Item.Hide()
	i.Visible = false
}

// SetVisible set whether current Item should be Visible
func (i *Item) SetVisible(visible bool) {
	switch visible {
	case true:
		i.Show()
	case false:
		i.Hide()
	}
}

// Refresh username and game name
func (i *Item) Refresh(s *twitch.Stream) {
	// sometimes the Twitch API does not send the username at first call, use the user UserLogin instead
	username := s.UserName
	if username == "" {
		username = s.UserLogin
	}

	i.Username = username
	i.Game = s.GameName
	i.Item.SetTitle(fmt.Sprintf("%s (%s)", i.Username, i.Game))
	i.Item.SetTooltip(s.Title)
}

func (i *Item) Disable() {
	i.Item.Disable()
}

func (i *Item) Click(ctx context.Context) {
	log.Debugf("starting click routine for [%s]", i.UserLogin)
	for {
		select {
		case <-ctx.Done():
			log.Debugf("received context cancel: Click [%s]", i.UserLogin)
			return // returning not to leak the goroutine
		case <-i.Item.ClickedCh:
			log.Debugf("[%s] Item is clicked", i.UserLogin)

			// Get link
			data, err := i.Application.Streamlink.Run(i.UserLogin)
			if err != nil {
				log.Errorln(err)
				continue // do not stop this routine in case of error
			}

			// Setting in clipboard
			u := strings.TrimSpace(string(data))
			i.Application.ClipboardListener <- u

			// Open in player and capture command output
			var out bytes.Buffer
			log.Debugf("openning with iina for [%s]", i.UserLogin)
			if err := i.Application.Player.Run(u, i.UserLogin, &out); err != nil {
				log.Errorf("[%s] cannot run command, received output: %s", i.Application.Player.Name(), out.String())
				continue // do not stop this routine in case of error
			}
		}
	}
}

// SetIcon pull avatar and set to given menu Item
func (i *Item) SetIcon() {
	users, err := i.Application.Twitch.Users.Get(i.UserLogin)
	if err != nil {
		log.Errorf("unable to refresh Twitch user info for %s: %s", i.UserLogin, err)
		return
	}

	if len(users) == 0 {
		log.Warningf("no image found for %s", i.UserLogin)
		return
	}

	// get user icon
	img, err := i.Application.Twitch.Users.ProfileImageBytes(users[0])
	if err != nil {
		log.Errorln(err)
		return
	}

	// set icon
	i.Item.SetIcon(img)
}

// ShouldNotify send true if current item is configured to send notifications
func (i *Item) ShouldNotify() bool {
	for _, user := range i.Application.config.Notifications {
		if i.UserLogin == strings.ToLower(user) {
			return true
		}
	}
	return false
}
