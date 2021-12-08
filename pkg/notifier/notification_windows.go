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

// startServer starts a web server to receive click callback events
func (s *service) startServer() {
	var port int
	once.Do(func() {
		port, err = freeport.GetFreePort()
		if err != nil {
			log.Errorf("notification service: cannot find free port, notifications click will not works: %s", err)
			return
		}
	})

	log.Tracef("notification service: using port %d", port)
	http.HandleFunc(actionURI, s.handleNotificationClick)
	s.srv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", serverListenAddr, port),
		Handler: nil, // let use http.DefaultServeMux
	}

	// start server
	go func() {
		log.Debugf("notification service: starting web server for notification click callblack at %s", s.srv.Addr)
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("notification service: fail to start web server: %s", err)
		}
	}()
}

// handleNotificationClick receives notification click events and
// send the streamer name to the service channel output
func (s *service) handleNotificationClick(_ http.ResponseWriter, r *http.Request) {
	stream := r.URL.Query().Get(streamQueryParameter)
	if stream == "" {
		log.Warningf("received notification callback without '%s' key", streamQueryParameter)
		return
	}

	log.Tracef("notification service: received notification event [%s]", stream)
	s.out <- stream
}
