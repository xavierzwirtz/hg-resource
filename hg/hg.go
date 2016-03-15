package hg

import (
	"os/exec"
	"bytes"
	"fmt"
	"strings"
)

type Repository struct {
	Path         string
	Branch       string
	IncludePaths []string
	ExcludePaths []string
	TagFilter    string
}

func (self *Repository) GetLatestCommitId() (string, error) {
	include := self.makeIncludeQueryFragment()
	exclude := self.makeExcludeQueryFragment()
	tagFilter := self.maybeTagFilter()
	revSet := fmt.Sprintf("last((((%s) - (%s)) & %s) - desc('[ci skip]'))", include, exclude, tagFilter)

	_, stdout, stderr, err := runHg([]string{
		"log",
		"--cwd", self.Path,
		"--rev", revSet,
		"--template", "{node}",
	})

	if err != nil {
		return "", fmt.Errorf("Error getting latest commit id: %s\nStderr: %s", err, stderr)
	}
	return stdout, nil
}

func (self *Repository) GetDescendantsOf(commitId string) ([]string, error) {
	include := self.makeIncludeQueryFragment()
	exclude := self.makeExcludeQueryFragment()
	tagFilter := self.maybeTagFilter()
	revSet := fmt.Sprintf("(descendants(%s) - %s) & %s & ((%s) - (%s)) - desc('[ci skip]')",
		commitId, commitId, tagFilter, include, exclude)

	_, stdout, stderr, err := runHg([]string{
		"log",
		"--cwd", self.Path,
		"--rev", revSet,
		"--template", "{node}\n",
	})

	if err != nil {
		return []string{}, fmt.Errorf("Error getting descendant commits of %s: %s\nStderr: %s", commitId, err, stderr)
	}

	trimmed := strings.Trim(stdout, "\n\r ")
	if len(trimmed) == 0 {
		return []string{}, nil
	}

	commits := strings.Split(trimmed, "\n")
	return commits, nil
}

func (self *Repository) maybeTagFilter() string {
	if len(self.TagFilter) > 0 {
		return "tag('re:" + escapePath(self.TagFilter) + "')"
	} else {
		return "all()"
	}
}

func runHg(args []string) (cmd *exec.Cmd, stdout string, stderr string, err error) {
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd = exec.Command("hg", args...)
	cmd.Stdout = outBuf
	cmd.Stderr = errBuf

	err = cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()
	return
}

func (self *Repository) makeIncludeQueryFragment() string {
	if len(self.IncludePaths) == 0 {
		return "all()"
	} else {
		return unionOfPaths(self.IncludePaths)
	}
}

func (self *Repository) makeExcludeQueryFragment() string {
	if len(self.ExcludePaths) == 0 {
		return "not all()"
	} else {
		return unionOfPaths(self.ExcludePaths)
	}
}

func unionOfPaths(paths []string) string {
	escapedPaths := make([]string, len(paths))
	for i, path := range (paths) {
		escapedPaths[i] = "file('re:" + escapePath(path) + "')"
	}
	return strings.Join(escapedPaths, "|")
}

func escapePath(path string) string {
	backslashesEscaped := strings.Replace(path, "\\", "\\\\\\\\", -1)
	quotesEscaped := strings.Replace(backslashesEscaped, "'", "\\'", -1)
	return quotesEscaped
}