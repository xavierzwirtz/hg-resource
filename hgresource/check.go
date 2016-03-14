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
	Source Source `json:"source"`
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
	}
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