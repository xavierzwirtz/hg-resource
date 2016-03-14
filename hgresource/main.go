// Concourse Mercurial Resource
// Compiles to a single multi-call binary, similar to BusyBox
// This file dispatches to {check, in, out}.go

package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

const usageTemplate = "Usage: hgresource <%s> [arguments]"

type Command struct {
	Name string
	Run  func([]string, io.Reader, io.Writer, io.Writer) int
}

var commands = []*Command{
	cmdCheck,
	cmdIn,
	cmdOut,
}

func main() {
	status := run(os.Args, os.Stdin, os.Stdout, os.Stderr)
	os.Exit(status)
}

func run(args []string, inReader io.Reader, outWriter io.Writer, errWriter io.Writer) int {
	// first, try to dispatch by application name
	appName := path.Base(args[0])
	handler, err := getHandler(appName, args[1:], outWriter, errWriter)
	if err == nil {
		return handler(args, inReader, outWriter, errWriter)
	}

	// then, check the first argument
	if len(args) > 1 {
		handler, err = getHandler(args[1], args[2:], outWriter, errWriter)
		if err == nil {
			argsCopy := []string{fmt.Sprintf("%s %s", path.Base(args[0]), args[1])}
			argsCopy = append(argsCopy, args[2:]...)
			return handler(argsCopy, inReader, outWriter, errWriter)
		}
	}

	usage(errWriter)
	return 2
}

func getHandler(name string, args []string, outWriter io.Writer, errWriter io.Writer) (func([]string, io.Reader, io.Writer, io.Writer) int, error) {
	for _, cmd := range (commands) {
		if cmd.Name == name {
			return cmd.Run, nil
		}
	}
	return nil, fmt.Errorf("command '%s' not found", name)
}

func makeUsage() string {
	var commandNames []string
	for _, cmd := range (commands) {
		commandNames = append(commandNames, cmd.Name)
	}

	joinedCommands := strings.Join(commandNames, "|")
	return fmt.Sprintf(usageTemplate, joinedCommands)
}

func usage(errWriter io.Writer) {
	errWriter.Write([]byte(makeUsage()))
}
