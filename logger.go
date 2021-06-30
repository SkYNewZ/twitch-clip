// +build production

package main

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

const logFileName = "run.log"

// Will configure log.SetOutput with the default user-specific logs directory, else /var/log/syslog
// Logs, if as files, will be rotated
func init() {
	log.SetLevel(log.TraceLevel) // default to trace level

	// get the system user log directory
	logDir, err := UserLogDir()
	log.Traceln("UserLogDir:", err)

	switch err == nil {
	case true:
		// We have the the system user log directory, use it
		output := configureRotateLogger(logDir, AppDisplayName)
		log.Printf("using log file: %s", output.Filename)
		log.SetOutput(output)
	case false:
		// Error, use the system log
		configureSyslogLogger()
		log.Println("using syslog")
	}
}

// configureRotateLogger return a lumberjack.Logger to use as log.SetOutput
func configureRotateLogger(path string, name string) *lumberjack.Logger {
	// Make a folder into if not exist
	appLogDir := filepath.Join(path, name)
	if _, err := os.Stat(appLogDir); errors.Is(err, os.ErrNotExist) {
		_ = os.Mkdir(appLogDir, 0755)
	}

	return &lumberjack.Logger{
		Filename:   filepath.Join(appLogDir, logFileName),
		MaxSize:    5, // megabytes
		MaxBackups: 2,
		MaxAge:     2, //days
		LocalTime:  true,
		Compress:   false, // disabled by default
	}
}

// UserLogDir returns the default root directory to use for user-specific
// logs file. Users should create their own application-specific subdirectory
// within this one and use that.
//
// On Unix systems, it returns $XDG_CACHE_HOME as specified by
// https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html if
// non-empty, else $HOME/.logs.
// On Darwin, it returns $HOME/Library/Logs.
// On Windows, it returns %LocalAppData%.
// On Plan 9, it returns $home/lib/logs.
//
// If the location cannot be determined (for example, $HOME is not defined),
// then it will return an error.
func UserLogDir() (string, error) {
	var dir string

	switch runtime.GOOS {
	case "windows":
		dir = os.Getenv("LocalAppData")
		if dir == "" {
			return "", errors.New("%LocalAppData% is not defined")
		}

	case "darwin", "ios":
		dir = os.Getenv("HOME")
		if dir == "" {
			return "", errors.New("$HOME is not defined")
		}
		dir += "/Library/Logs"

	case "plan9":
		dir = os.Getenv("home")
		if dir == "" {
			return "", errors.New("$home is not defined")
		}
		dir += "/lib/logs"

	default: // Unix
		dir = os.Getenv("XDG_STATE_HOME")
		if dir == "" {
			dir = os.Getenv("HOME")
			if dir == "" {
				return "", errors.New("neither $XDG_STATE_HOME nor $HOME are defined")
			}
			dir += "/.logs"
		}
	}

	return dir, nil
}
