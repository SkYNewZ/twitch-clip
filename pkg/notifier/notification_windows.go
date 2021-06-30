// +build windows

package notifier

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/stuartleeks/toast"
)

func (s *service) Notify(username, game, id string) error {
	log.Tracef("notification service: creating notification for [%s]", username)
	notification := toast.Notification{
		AppID:    s.title,
		Title:    defaultTitle,
		Message:  fmt.Sprintf(defaultSubtitle, username, game),
		Icon:     "C:\\Users\\Quentin\\Sources\\alerts\\assets\\icon256.png",
		Actions:  nil,
		Audio:    toast.Default,
		Loop:     false,
		Duration: "",
	}

	// if local server started, append click action
	if s.srv != nil {
		notification.Actions = []toast.Action{
			{
				Type:      "protocol",
				Label:     "View",
				Arguments: s.makeNotificationURL(id),
			},
		}
	}

	return notification.Push()
}
