package notifier

import (
	"fmt"

	"github.com/gen2brain/beeep"
)

func (s *service) Notify(username, game, _ string) error {
	return beeep.Notify(s.title, fmt.Sprintf(defaultSubtitle, username, game), "")
}
