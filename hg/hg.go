package hg

import (
	"os/exec"
	"bytes"
	"fmt"
	"strings"
	"os"
)

type Repository struct {
	Path         string
	Branch       string
	IncludePaths []string
	ExcludePaths []string
	TagFilter    string
}

func (self *Repository) CloneOrPull(sourceUri string, insecure bool) error {
	if len(self.Path) == 0 {
		return fmt.Errorf("CloneOrPull: repository path must be set")
	}

	if len(self.Branch) == 0 {
		return fmt.Errorf("CloneOrPull: branch must be set")
	}

	dirInfo, errIfNotExists := os.Stat(self.Path)
	if errIfNotExists != nil || !dirInfo.IsDir() {
		return self.clone(sourceUri, insecure)
	} else {
		return self.pull(insecure)
	}
}

func (self *Repository) clone(sourceUri string, insecure bool) error {
	err := os.RemoveAll(self.Path)
	if err != nil {
		return fmt.Errorf("CloneOrUpdate: %s", err)
	}

	_, _, stderr, err := runHg("clone", []string{
		"-q",
		"--branch", self.Branch,
		sourceUri,
		self.Path,
	}, insecure)
	if err != nil {
		return fmt.Errorf("Error cloning repository from %s: %s\nStderr: %s", sourceUri, err, stderr)
	}

	return nil
}

func (self *Repository) pull(insecure bool) error {
	_, _, stderr, err := runHg("pull", []string{
		"-q",
		"--cwd", self.Path,
	}, insecure)
	if err != nil {
		return fmt.Errorf("Error pulling changes from repository: %s\nStderr: %s", err, stderr)
	}

	_, _, stderr, err = runHg("checkout", []string{
		"-q",
		"--cwd", self.Path,
		"--clean",
		"--rev", "tip",
	}, insecure)
	if err != nil {
		return fmt.Errorf("Error updating working directory to tip: %s\nStderr: %s", err, stderr)
	}

	return nil
}

func (self *Repository) GetLatestCommitId() (string, error) {
	include := self.makeIncludeQueryFragment()
	exclude := self.makeExcludeQueryFragment()
	tagFilter := self.maybeTagFilter()
	revSet := fmt.Sprintf("last((((%s) - (%s)) & %s) - desc('[ci skip]'))", include, exclude, tagFilter)

	_, stdout, stderr, err := runHg("log", []string{
		"--cwd", self.Path,
		"--rev", revSet,
		"--template", "{node}",
	}, false)

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

	_, stdout, stderr, err := runHg("log", []string{
		"--cwd", self.Path,
		"--rev", revSet,
		"--template", "{node}\n",
	}, false)

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

func runHg(command string, args []string, insecure bool) (cmd *exec.Cmd, stdout string, stderr string, err error) {
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	hgArgs := make([]string, 1, len(args) + 1)
	hgArgs[0] = command
	if insecure && commandTakesInsecureOption(command) {
		hgArgs = append(hgArgs, "--insecure")
	}
	hgArgs = append(hgArgs, args...)

	cmd = exec.Command("hg", hgArgs...)
	cmd.Stdout = outBuf
	cmd.Stderr = errBuf

	err = cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()
	return
}

func commandTakesInsecureOption(command string) bool {
	eligibleCommands := []string{
		"clone",
		"pull",
		"push",
	}
	for _, eligibleCommand := range (eligibleCommands) {
		if command == eligibleCommand {
			return true
		}
	}
	return false
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