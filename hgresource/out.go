package main

import "io"

var cmdOutName string = "out"
var cmdOut = &Command{
	Name: cmdOutName,
	Run: runOut,
}

func runOut(args []string, inReader io.Reader, outWriter io.Writer, errWriter io.Writer) int {
	errWriter.Write([]byte("Not implemented"))
	return 1
}
