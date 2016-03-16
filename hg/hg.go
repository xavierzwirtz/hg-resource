package hg

import (
	"os/exec"
	"fmt"
	"strings"
	"os"
)

type Repository struct {
	Path                string
	Branch              string
	IncludePaths        []string
	ExcludePaths        []string
	TagFilter           string
	SkipSslVerification bool
}

type HgMetadata struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

func (self *Repository) CloneOrPull(sourceUri string, insecure bool) ([]byte, error) {
	if len(self.Path) == 0 {
		return []byte{}, fmt.Errorf("CloneOrPull: repository path must be set")
	}

	if len(self.Branch) == 0 {
		return []byte{}, fmt.Errorf("CloneOrPull: branch must be set")
	}

	dirInfo, errIfNotExists := os.Stat(self.Path)
	if errIfNotExists != nil || !dirInfo.IsDir() {
		return self.clone(sourceUri, insecure)
	} else {
		return self.pull(insecure)
	}
}

func (self *Repository) clone(sourceUri string, insecure bool) (output []byte, err error) {
	err = os.RemoveAll(self.Path)
	if err != nil {
		err = fmt.Errorf("CloneOrUpdate: %s", err)
		return
	}

	_, output, err = self.run("clone", []string{
		"-q",
		"--branch", self.Branch,
		sourceUri,
		self.Path,
	})
	if err != nil {
		err = fmt.Errorf("Error cloning repository from %s: %s", sourceUri, err)
	}

	return
}

func (self *Repository) pull(insecure bool) (output []byte, err error) {
	_, output, err = self.run("pull", []string{
		"-q",
		"--cwd", self.Path,
	})
	if err != nil {
		err = fmt.Errorf("Error pulling changes from repository: %s\nStderr", err)
		return
	}

	_, checkoutOutput, err := self.run("checkout", []string{
		"-q",
		"--cwd", self.Path,
		"--clean",
		"--rev", "tip",
	})
	output = append(output, checkoutOutput...)
	if err != nil {
		err = fmt.Errorf("Error updating working directory to tip: %s", err)
	}

	return
}

func (self *Repository) PullWithRebase(sourceUri string, branch string) (output []byte, err error) {
	_, output, err = self.run("pull", []string{
		"-q",
		"--cwd", self.Path,
		"--config", "extensions.rebase=",
		"--config", "paths.push-target=" + sourceUri,
		"--rebase",
		"--branch", branch,
		"push-target",
	})
	if err != nil {
		err = fmt.Errorf("Error pulling/rebasing from: %s: %s", sourceUri, err)
	}
	return
}

func (self *Repository) CloneAtCommit(sourceUri string, commitId string) (output []byte, err error) {
	_, output, err = self.run("clone", []string{
		"-q",
		"--rev", commitId,
		sourceUri,
		self.Path,
	})
	if err != nil {
		err = fmt.Errorf("Error cloning repository %s@%s: %s", sourceUri, commitId, err)
	}

	return
}

func (self *Repository) SetDraftPhase() (output []byte, err error) {
	_, output, err = self.run("phase", []string{
		"--cwd", self.Path,
		"--force",
		"--draft",
	})
	if err != nil {
		err = fmt.Errorf("Error setting repo phase to draft: %s", err)
	}

	return
}

func (self *Repository) Push(destUri string, branch string) (output []byte, err error) {
	_, output, err = self.run("push", []string{
		"--cwd", self.Path,
		"--config", "paths.push-target=" + destUri,
		"--branch", branch,
		"push-target",
	})
	if err != nil {
		err = fmt.Errorf("Error pushing to %s: %s", destUri, err)
	}

	return
}

func (self *Repository) Tag(tagValue string) (output []byte, err error) {
	_, output, err = self.run("tag", []string{
		"--cwd", self.Path,
		tagValue,
	})
	if err != nil {
		err = fmt.Errorf("Error tagging current commit: %s", err)
	}

	return
}

func (self *Repository) Delete() error {
	err := os.RemoveAll(self.Path)
	if err != nil {
		return fmt.Errorf("Error deleting repository: %s", err)
	}

	return nil
}

func (self *Repository) Checkout(commitId string) (output []byte, err error) {
	_, output, err = self.run("checkout", []string{
		"-q",
		"--cwd", self.Path,
		"--clean",
		"--rev", commitId,
	})
	if err != nil {
		err = fmt.Errorf("Error checking out %s: %s", commitId, err)
	}

	return
}

func (self *Repository) Purge() (output []byte, err error) {
	_, output, err = self.run("purge", []string{
		"--config", "extensions.purge=",
		"--cwd", self.Path,
		"--all",
	})
	if err != nil {
		err = fmt.Errorf("Error purging repository: %s", err)
	}

	return
}

func (self *Repository) GetLatestCommitId() (output string, err error) {
	include := self.makeIncludeQueryFragment()
	exclude := self.makeExcludeQueryFragment()
	tagFilter := self.maybeTagFilter()
	revSet := fmt.Sprintf("last((((%s) - (%s)) & %s) - desc('[ci skip]'))", include, exclude, tagFilter)

	_, outBytes, err := self.run("log", []string{
		"--cwd", self.Path,
		"--rev", revSet,
		"--template", "{node}",
	})
	output = string(outBytes)
	if err != nil {
		err = fmt.Errorf("Error getting latest commit id: %s", err)
	}

	return
}

func (self *Repository) GetCurrentCommitId() (output string, err error) {
	_, outBytes, err := self.run("log", []string{
		"--cwd", self.Path,
		"--rev", ".",
		"--template", "{node}",
	})
	output = string(outBytes)
	if err != nil {
		err = fmt.Errorf("Error getting current commit id: %s", err)
	}

	return
}

func (self *Repository) GetDescendantsOf(commitId string) ([]string, error) {
	include := self.makeIncludeQueryFragment()
	exclude := self.makeExcludeQueryFragment()
	tagFilter := self.maybeTagFilter()
	revSet := fmt.Sprintf("(descendants(%s) - %s) & %s & ((%s) - (%s)) - desc('[ci skip]')",
		commitId, commitId, tagFilter, include, exclude)

	_, outBytes, err := self.run("log", []string{
		"--cwd", self.Path,
		"--rev", revSet,
		"--template", "{node}\n",
	})

	output := string(outBytes)
	if err != nil {
		return []string{}, fmt.Errorf("Error getting descendant commits of %s: %s\n%s", commitId, err, output)
	}

	trimmed := strings.Trim(output, "\n\r ")
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

func (self *Repository) Metadata(commitId string) (fullCommitId string, metadata []HgMetadata, err error) {
	// TODO use single call to hg, e.g. via json template + date conversion in go
	// TODO date format in json is: [epochSeconds, secondOffsetFromUTC]
	_, outBytes, err := self.run("log", []string{
		"--cwd", self.Path,
		"--rev", commitId,
		"--template", "{node}",
	})
	fullCommitId = string(outBytes)
	if err != nil {
		err = fmt.Errorf("Error getting metadata on %s: %s\n%s", commitId, err, fullCommitId)
		return
	}

	_, outBytes, err = self.run("log", []string{
		"--cwd", self.Path,
		"--rev", commitId,
		"--template", "{author}",
	})
	author := string(outBytes)
	if err != nil {
		err = fmt.Errorf("Error getting metadata on %s: %s\n%s", commitId, err, author)
		return
	}

	_, outBytes, err = self.run("log", []string{
		"--cwd", self.Path,
		"--rev", commitId,
		"--template", "{date|isodatesec}",
	})
	date := string(outBytes)
	if err != nil {
		err = fmt.Errorf("Error getting metadata on %s: %s\n%s", commitId, err, date)
		return
	}

	_, outBytes, err = self.run("log", []string{
		"--cwd", self.Path,
		"--rev", commitId,
		"--template", "{desc}",
	})
	message := string(outBytes)
	if err != nil {
		err = fmt.Errorf("Error getting metadata on %s: %s\n%s", commitId, err, message)
		return
	}

	metadata = append(metadata,
		HgMetadata{
			Name: "commit",
			Value: fullCommitId,
		},
		HgMetadata{
			Name: "author",
			Value: author,
		},
		HgMetadata{
			Name: "author_date",
			Value: date,
			Type: "time",
		},
		HgMetadata{
			Name: "message",
			Value: message,
			Type: "message",
		},
	)
	return
}

func (self *Repository) run(command string, args []string) (cmd *exec.Cmd, output []byte, err error) {
	hgArgs := make([]string, 1, len(args) + 1)
	hgArgs[0] = command

	if self.SkipSslVerification && commandTakesInsecureOption(command) {
		hgArgs = append(hgArgs, "--insecure")
	}
	hgArgs = append(hgArgs, args...)

	cmd = exec.Command("hg", hgArgs...)

	output, err = cmd.CombinedOutput()
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