//go:build !windows
// +build !windows

package player

func (p *player) checkRegistry() bool {
	return false
}
