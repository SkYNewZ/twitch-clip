package notifier

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/phayes/freeport"
	log "github.com/sirupsen/logrus"
)

const (
	defaultTitle         = "A stream has just started"
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

	once.Do(func() {
		port, err := freeport.GetFreePort()
		if err != nil {
			log.Errorf("notification service: cannot find free port, notifications click will not works: %s", err)
			return
		}

		s.startServer(port)
	})

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

func (s *service) startServer(port int) {
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

func (s *service) makeNotificationURL(streamer string) string {
	u, _ := url.Parse("http://" + s.srv.Addr + actionURI)
	q := u.Query()
	q.Set(streamQueryParameter, streamer)
	u.RawQuery = q.Encode()
	return u.String()
}
