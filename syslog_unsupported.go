//go:build production && (windows || nacl || plan9)

package main

func configureSyslogLogger() {}
