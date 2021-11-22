//go:build production && (windows || nacl || plan9)
// +build production
// +build windows nacl plan9

package main

func configureSyslogLogger() {}
