package notifier

import (
	"fmt"

	"github.com/gen2brain/beeep"
)

func (s *service) Notify(username, game, _ string) error {
	return beeep.Notify(s.title, fmt.Sprintf(defaultSubtitle, username, game), "")
}

// startServer notification callback handler is not supported on darwin as it runs a AppleScript
func (s *service) startServer() {}
