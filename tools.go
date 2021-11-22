//go:build tools
// +build tools

package main

import (
	_ "github.com/cratonica/2goarray"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/tc-hib/go-winres"
)
