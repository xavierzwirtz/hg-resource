package main

import (
	"io"
	"github.com/andreasf/hg-resource/hg"
	"bytes"
	"encoding/json"
	"fmt"
	"path"
)

const (
	defaultBranch = "default"
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
		Path: getCacheDir(),
		Branch: params.Source.Branch,
		IncludePaths: params.Source.IncludePaths,
		ExcludePaths: params.Source.ExcludePaths,
		TagFilter: params.Source.TagFilter,
	}

	if len(repo.Branch) == 0 {
		repo.Branch = defaultBranch
	}

	if len(params.Source.PrivateKey) != 0 {
		err = loadSshPrivateKey(params.Source.PrivateKey)
		if err != nil {
			fmt.Fprintln(errWriter, err)
			return 1
		}
	}

	if len(params.Source.Uri) == 0 {
		fmt.Fprintln(errWriter, "Repository URI must be provided")
		return 1
	}

	err = repo.CloneOrPull(params.Source.Uri, params.Source.SkipSslVerification)
	if err != nil {
		fmt.Fprintln(errWriter, err)
		return 1
	}

	if len(params.Version.Ref) == 0 {
		return writeLatestCommit(&repo, outWriter, errWriter)
	} else {
		return writeCommitsSince(params.Version.Ref, &repo, outWriter, errWriter)
	}
}

func getCacheDir() string {
	return path.Join(getTempDir(), "hg-resource-repo-cache")
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

func writeCommitsSince(parentCommit string, repo *hg.Repository, outWriter io.Writer, errWriter io.Writer) int {
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