package main

import (
	"fmt"
	"io"
	"github.com/andreasf/hg-resource/hg"
)

var cmdInName string = "in"
var cmdIn = &Command{
	Name: cmdInName,
	Run: runIn,
}

func runIn(args []string, inReader io.Reader, outWriter io.Writer, errWriter io.Writer) int {
	if len(args) < 2 {
		inUsage(args[0], errWriter)
		return 2
	}

	destination := args[1]
	params, err := parseInput(inReader)

	if err != nil {
		fmt.Fprintf(errWriter, "Error parsing input: %s\n", err)
		return 1
	}

	repo := &hg.Repository{
		Path: destination,
		Branch: params.Source.Branch,
		IncludePaths: params.Source.IncludePaths,
		ExcludePaths: params.Source.ExcludePaths,
		TagFilter: params.Source.TagFilter,
		SkipSslVerification: params.Source.SkipSslVerification,
	}

	if len(repo.Branch) == 0 {
		repo.Branch = defaultBranch
	}

	if len(params.Source.Uri) == 0 {
		fmt.Fprintln(errWriter, "Repository URI must be provided")
		return 1
	}

	var commitId string
	if len(params.Version.Ref) == 0 {
		commitId = "tip"
	} else {
		commitId = params.Version.Ref
	}

	if len(params.Source.PrivateKey) != 0 {
		err = loadSshPrivateKey(params.Source.PrivateKey)
		if err != nil {
			fmt.Fprintln(errWriter, err)
			return 1
		}
	}

	output, err := repo.CloneOrPull(params.Source.Uri, params.Source.SkipSslVerification)
	errWriter.Write(output)
	if err != nil {
		fmt.Fprintln(errWriter, err)
		return 1
	}

	output, err = repo.Checkout(commitId)
	errWriter.Write(output)
	if err != nil {
		fmt.Fprintln(errWriter, err)
		return 1
	}

	output, err = repo.Purge()
	errWriter.Write(output)
	if err != nil {
		fmt.Fprintln(errWriter, err)
		return 1
	}

	jsonOutput, err := getJsonOutputForCurrentCommit(repo)
	if err != nil {
		fmt.Fprintln(errWriter, err)
		return 1
	}
	WriteJson(outWriter, jsonOutput)
	return 0
}

func inUsage(appName string, err io.Writer) {
	errMsg := fmt.Sprintf("Usage: %s <path/to/destination>", appName)
	err.Write([]byte(errMsg))
}