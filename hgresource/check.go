package main

import (
	"io"
	"github.com/andreasf/hg-resource/hg"
	"bytes"
	"encoding/json"
	"fmt"
)

var cmdCheckName string = "check"
var cmdCheck = &Command{
	Name: cmdCheckName,
	Run: runCheck,
}

type InParams struct {
	Source  Source `json:"source"`
	Version Version `json:"version"`
}

type VersionList []Version

func runCheck(args []string, inReader io.Reader, outWriter io.Writer, errWriter io.Writer) int {
	params, err := parseInput(inReader)
	if err != nil {
		fmt.Fprintf(errWriter, "Error parsing input: %s\n", err)
		return 1
	}

	repo := hg.Repository{
		Path: params.Source.Uri,
		Branch: params.Source.Branch,
		IncludePaths: params.Source.IncludePaths,
		ExcludePaths: params.Source.ExcludePaths,
		TagFilter: params.Source.TagFilter,
	}

	switch true {
	case params.Source.Uri == "":
		fmt.Fprintln(errWriter, "Repository URI must be provided")
		return 1
	case params.Version.Ref == "" && params.Source.Uri != "":
		return writeLatestCommit(&repo, outWriter, errWriter)
	case params.Version.Ref != "" && params.Source.Uri != "":
		return WriteCommitsSince(params.Version.Ref, &repo, outWriter, errWriter)
	default:
		panic("Unreachable statement")
	}
}

func writeLatestCommit(repo *hg.Repository, outWriter io.Writer, errWriter io.Writer) int {
	latestCommit, err := repo.GetLatestCommitId()
	if err != nil {
		fmt.Fprintln(errWriter, err)
		return 1
	}

	latestVersion := VersionList{
		Version{
			Ref: latestCommit,
		},
	}

	_, err = WriteJson(outWriter, latestVersion)
	if err != nil {
		fmt.Fprintln(errWriter, err)
		return 1
	}
	return 0
}

func WriteCommitsSince(parentCommit string, repo *hg.Repository, outWriter io.Writer, errWriter io.Writer) int {
	commits, err := repo.GetDescendantsOf(parentCommit)
	if err != nil {
		// commit id not found -- return latest commit as fallback
		return writeLatestCommit(repo, outWriter, errWriter)
	}

	commitList := make(VersionList, len(commits))
	for i, commit := range (commits) {
		commitList[i] = Version{
			Ref: commit,
		}
	}

	_, err = WriteJson(outWriter, commitList)
	if err != nil {
		fmt.Fprintln(errWriter, err)
		return 1
	}

	return 0
}

func parseInput(inReader io.Reader) (*InParams, error) {
	bytes, err := readAllBytes(inReader)
	if err != nil {
		return nil, err
	}

	params := InParams{}
	json.Unmarshal(bytes, &params)
	return &params, nil
}

func readAllBytes(reader io.Reader) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)

	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func WriteJson(outWriter io.Writer, object interface{}) (n int, err error) {
	output, err := json.Marshal(object)
	if err != nil {
		return n, fmt.Errorf("Error serializing JSON response: %s", err)
	}

	n, err = outWriter.Write(output)
	if err != nil {
		return n, fmt.Errorf("Error writing JSON to io.Writer: %s", err)
	}

	outWriter.Write([]byte("\n"))
	n++
	return
}