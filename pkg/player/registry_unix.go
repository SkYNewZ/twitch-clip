// +build !windows

package player

func (p *player) checkRegistry() bool {
	_ = p.registry
	_ = p.registry32
	return false
}
