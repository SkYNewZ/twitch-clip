package twitch

import "fmt"

var _ error = (*Error)(nil)

// Error describes a Twitch error
type Error struct {
	Err     string `json:"error"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return fmt.Sprintf("twitch error %d %s: %s", e.Status, e.Err, e.Message)
}
