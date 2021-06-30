// +build !windows

package notifier

import (
	"errors"
	"runtime"
)

var ErrUnsupported = errors.New("notification service: unsupported operation system: " + runtime.GOOS)

func (s *service) Notify(username, game, id string) error {
	return ErrUnsupported
}
