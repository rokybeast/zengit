package git

import (
	"fmt"
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

// check status of repo changes and pushes
func CheckRepoStatus() (hasChanges bool, hasPushes bool) {
	// run git status porcelain to see if working tree is dirty
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusOut, err := statusCmd.Output()
	if err == nil && len(strings.TrimSpace(string(statusOut))) > 0 {
		hasChanges = true
	}

	// check if branch is ahead of upstream [gitdocs]
	upstreamCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "@{u}")
	upstreamOut, err := upstreamCmd.Output()
	if err == nil {
		upstream := strings.TrimSpace(string(upstreamOut))
		countCmd := exec.Command("git", "rev-list", "--count", upstream+"..HEAD")
		countOut, err := countCmd.Output()
		if err == nil {
			var count int
			_, _ = fmt.Sscanf(strings.TrimSpace(string(countOut)), "%d", &count)
			if count > 0 {
				hasPushes = true
				return
			}
		}
	} else {
		// if no upstream, check if there are any remotes [gitdocs]
		remoteCmd := exec.Command("git", "remote")
		remoteOut, _ := remoteCmd.Output()
		if len(strings.TrimSpace(string(remoteOut))) > 0 {
			// there is at least one remote, check if we have commits not pushed to any remote
			countCmd := exec.Command("git", "rev-list", "--count", "HEAD", "--not", "--remotes")
			countOut, err := countCmd.Output()
			if err == nil {
				var count int
				_, _ = fmt.Sscanf(strings.TrimSpace(string(countOut)), "%d", &count)
				if count > 0 {
					hasPushes = true
					return
				}
			}
		}
	}

	return
}

// get the latest short commit sha
func LatestShortSHA() string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// check if there are any staged files ready for commit
func HasStagedFiles() bool {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

