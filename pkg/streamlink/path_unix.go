//go:build !windows
// +build !windows

package streamlink

import "os"

func init() {
	// Append some paths in PATH (https://stackoverflow.com/questions/27451697/not-able-to-execute-go-file-using-os-exec-package)
	origin := os.Getenv("PATH")
	_ = os.Setenv("PATH", "/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/sbin:"+origin)
}
