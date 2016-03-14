package main

import "io"

var cmdCheckName string = "check"
var cmdCheck = &Command{
	Name: cmdCheckName,
	Run: runCheck,
}

func runCheck(args []string, outWriter io.Writer, errWriter io.Writer) int {
	errWriter.Write([]byte("Not implemented"))
	return 1
}
