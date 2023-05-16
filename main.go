package main

import (
	"github.com/getlantern/systray"
	log "github.com/sirupsen/logrus"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered from panic: %s", r)
		}
	}()

	app := New()
	systray.Run(app.Setup, app.Stop)
}
