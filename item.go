package main

import (
	"bytes"
	"context"
	"strings"
	"sync"

	"github.com/getlantern/systray"
	log "github.com/sirupsen/logrus"
)

// Item describes a displayed item
type Item struct {
	Application *application
	Item        *systray.MenuItem
	Visible     bool
	ID          string
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

// Refresh refresh Item info
func (i *Item) Refresh(title, tooltip string) {
	i.Item.SetTitle(title)
	i.Item.SetTooltip(tooltip)
}

func (i *Item) Disable() {
	i.Item.Disable()
}

func (i *Item) Click(ctx context.Context) {
	log.Debugf("starting click routine for [%s]", i.ID)
	for {
		select {
		case <-ctx.Done():
			log.Debugf("received context cancel: Click [%s]", i.ID)
			return // returning not to leak the goroutine
		case <-i.Item.ClickedCh:
			log.Debugf("[%s] Item is clicked", i.ID)

			// Get link
			data, err := i.Application.Streamlink.Run(i.ID)
			u := strings.TrimSpace(string(data))
			if err != nil {
				log.Errorln(err)
				return
			}

			// Setting in clipboard
			i.Application.ClipboardListener <- u

			// Open in player and capture command output
			var out bytes.Buffer
			log.Debugf("openning with iina for [%s]", i.ID)
			if err := i.Application.Player.Run(u, i.ID, &out); err != nil {
				log.Errorf("[%s] cannot run command, received output: %s", i.Application.Player.Name(), out.String())
				return
			}
		}
	}
}

// SetIcon pull avatar and set to given menu Item
func (i *Item) SetIcon() {
	users, err := i.Application.Twitch.Users.Get(i.ID)
	if err != nil {
		log.Errorf("unable to refresh Twitch user info for %s: %s", i.ID, err)
		return
	}

	if len(users) == 0 {
		log.Warningf("no image found for %s", i.ID)
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
