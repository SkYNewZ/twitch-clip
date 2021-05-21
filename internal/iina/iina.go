package iina

import (
	"os/exec"

	log "github.com/sirupsen/logrus"
)

const iinaBin = "/usr/local/bin/iina"

var iinaArgs = []string{
	"--no-stdin",
}

// InPath check whether streamlink is in path
func InPath() bool {
	_, err := exec.LookPath(iinaBin)
	return err == nil
}

// Run runs given URL in IINA
func Run(link string) error {
	cmdArgs := iinaArgs
	cmdArgs = append(cmdArgs, link)

	cmd := exec.Command(iinaBin, cmdArgs...)
	log.Tracef("running command [%s]", cmd.String())
	return cmd.Run()
}
