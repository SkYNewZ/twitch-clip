// +build production
// +build !windows,!nacl,!plan9

package main

import (
	"log/syslog"

	log "github.com/sirupsen/logrus"
)

func configureSyslogLogger() {
	if hook, err := logrussyslog.NewSyslogHook("", "", syslog.LOG_DEBUG, ""); err == nil {
		log.AddHook(hook)
	}
}