package main

import (
	"fmt"
	"io"
)

var cmdInName string = "in"
var cmdIn = &Command{
	Name: cmdInName,
	Run: runIn,
}

func runIn(args []string, outWriter io.Writer, errWriter io.Writer) int {
	if len(args) < 2 {
		inUsage(args[0], errWriter)
		return 2
	}
	errWriter.Write([]byte("Not implemented"))
	return 1
}

func inUsage(appName string, err io.Writer) {
	errMsg := fmt.Sprintf("Usage: %s <path/to/destination>", appName)
	err.Write([]byte(errMsg))
}