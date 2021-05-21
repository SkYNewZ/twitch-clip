package main

import (
	"context"
	"os"
	"strings"

	"github.com/SkYNewZ/twitch-clip/internal/icon"
	"github.com/SkYNewZ/twitch-clip/internal/iina"

	"github.com/SkYNewZ/twitch-clip/internal/streamlink"
	"github.com/emersion/go-autostart"
	"github.com/getlantern/systray"
	log "github.com/sirupsen/logrus"
)

const (
	AppName        = "twitchclip"
	AppDisplayName = "Twitch Clip"
)

var (
	mainCancelFunc context.CancelFunc
)

func init() {
	log.SetLevel(log.TraceLevel) // default log level

	// streamlink in path ?
	if !streamlink.InPath() {
		log.Fatalln("streamlink missing in $PATH")
	}

	// iina in path ?
	if !iina.InPath() {
		log.Fatalln("iina missing in $PATH")
	}
}

func getStreamLink(name string) (string, error) {
	output, err := streamlink.ExecuteStreamlinkCommand(name)
	return strings.TrimSpace(string(output)), err
}

func main() {
	onExit := func() {
		mainCancelFunc()
	}

	systray.Run(onReady, onExit)
}

// onReady set up this app
func onReady() {
	executable, err := os.Executable()
	if err != nil {
		log.Fatalln(err)
	}
	app := &autostart.App{
		Name:        AppName,
		DisplayName: AppDisplayName,
		Exec:        []string{executable},
	}

	systray.SetIcon(icon.Data)
	systray.SetTooltip(AppDisplayName)

	// Start app on start
	startOnStartup := systray.AddMenuItemCheckbox("Start at login", "Start this app at system startup", app.IsEnabled())
	go toggleStartOnStartup(startOnStartup, app)

	// Quit menu
	quit := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		<-quit.ClickedCh
		systray.Quit()
	}()
	systray.AddSeparator()

	setupTwitch()

	// main context to stop routines
	var ctx context.Context
	ctx, mainCancelFunc = context.WithCancel(context.Background())
	setupStreamsMenuItem(ctx)
}

// toggleStartOnStartup manage if this app must start on system startup or not
func toggleStartOnStartup(item *systray.MenuItem, app *autostart.App) {
	for {
		<-item.ClickedCh // wait for a click
		switch app.IsEnabled() {
		case true:
			log.Debugln("disable application autostart")
			if err := app.Disable(); err != nil {
				log.Errorln(err)
				continue
			}

			item.Uncheck()
		case false:
			log.Debugln("enable application autostart")
			if err := app.Enable(); err != nil {
				log.Errorln(err)
				continue
			}

			item.Check()
		}
	}
}
