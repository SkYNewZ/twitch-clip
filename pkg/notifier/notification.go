package notifier

import (
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
)

const (
	defaultSubtitle      = "%s start streaming %s"
	actionURI            = "/notification" // Use URI not handle notifications callback
	serverListenAddr     = "localhost"
	streamQueryParameter = "id"
)

var once sync.Once

// Notifier service
type Notifier interface {
	// Notify send a desktop notification
	Notify(username, game, id string) error

	// Close stops the current notifier service (closes the underlying web server)
	Close() error
}

// service implements Notifier
type service struct {
	title string        // Notification application title
	out   chan<- string // send notifications click events
	srv   *http.Server  // server which handle notification click callbacks
}

// New creates a new notifier service and output channel for notification callback events
func New(title string) (Notifier, <-chan string) {
	out := make(chan string)
	var s = &service{
		title: title,
		out:   out,
		srv:   nil,
	}

	s.startServer()
	return s, out
}

func (s *service) Close() error {
	close(s.out)

	// server does not exist
	if s.srv == nil {
		return nil
	}

	log.Debugln("notification service: closing web server")
	if err := s.srv.Close(); err != nil {
		return err
	}

	return nil
}
