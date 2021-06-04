package main

import (
	"github.com/getlantern/systray"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

func main() {
	app := New()
	systray.Run(app.Setup, app.Stop)
}
