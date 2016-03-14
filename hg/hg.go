package hg

import (
	"os/exec"
	"bytes"
	"fmt"
	"strings"
)

type Repository struct {
	Path string
}

func (self *Repository) GetLatestCommitId() (string, error) {
	_, stdout, stderr, err := runHg([]string{
		"log",
		"--cwd", self.Path,
		"--rev", "tip",
		"--template", "{node}",
	})

	if err != nil {
		return "", fmt.Errorf("Error getting latest commit id: %s\nStderr: %s", err, stderr)
	}
	return stdout, nil
}

func (self *Repository) GetDescendantsOf(commitId string) ([]string, error) {
	revSet := fmt.Sprintf("descendants(%s) - %s", commitId, commitId)
	_, stdout, stderr, err := runHg([]string{
		"log",
		"--cwd", self.Path,
		"--rev", revSet,
		"--template", "{node}\n",
	})

	if err != nil {
		return []string{}, fmt.Errorf("Error getting descendant commits of %s: %s\nStderr: %s", commitId, err, stderr)
	}
	commits := strings.Split(strings.Trim(stdout, "\n"), "\n")
	return commits, nil
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