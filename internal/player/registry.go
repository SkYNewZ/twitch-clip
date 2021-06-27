// +build windows

package player

import (
	"runtime"

	"golang.org/x/sys/windows/registry"
)

// checkRegistry check whether given command exist in Windows registry
// https://github.com/SoMuchForSubtlety/f1viewer/blob/master/internal/cmd/registry.go
func (p *player) checkRegistry() bool {
	regPath := p.registry
	if runtime.GOARCH == "386" {
		regPath = p.registry32
	}

	if regPath == "" {
		return false
	}

	result, err := registry.OpenKey(registry.LOCAL_MACHINE, regPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}

	path, _, err := result.GetStringValue("InstallDir")
	if err != nil {
		return false
	}

	p.command[0] = path + "\\" + p.command[0]
	return true
}
