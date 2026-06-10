package git

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// get the repository name
func RepoName() string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	path := strings.TrimSpace(string(out))
	return filepath.Base(path)
}

// get the current branch name
func CurrentBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}
