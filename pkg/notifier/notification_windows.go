package notifier

import (
	"fmt"
	"net/url"

	log "github.com/sirupsen/logrus"
	"github.com/stuartleeks/toast"
)

func (s *service) Notify(username, game, id string) error {
	log.Tracef("notification service: creating notification for [%s]", username)
	notification := toast.Notification{
		AppID:    s.title,
		Title:    s.title,
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

func (s *service) makeNotificationURL(streamer string) string {
	u, _ := url.Parse("http://" + s.srv.Addr + actionURI)
	q := u.Query()
	q.Set(streamQueryParameter, streamer)
	u.RawQuery = q.Encode()
	return u.String()
}
