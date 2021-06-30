package main

import "github.com/getlantern/systray"

func main() {
	app := New()
	systray.Run(app.Setup, app.Stop)
}
