//go:build production && !windows && !nacl && !plan9

package main

import (
	"log/syslog"

	log "github.com/sirupsen/logrus"
	logrussyslog "github.com/sirupsen/logrus/hooks/syslog"
)

func configureSyslogLogger() {
	if hook, err := logrussyslog.NewSyslogHook("", "", syslog.LOG_DEBUG, ""); err == nil {
		log.AddHook(hook)
		log.Println("using syslog")
	}
}
