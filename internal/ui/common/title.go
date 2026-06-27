package common

import "fmt"

const Version = "v0.8.5"

func GetTitle(repoName, branch string) string {
	if repoName != "" && branch != "" {
		return fmt.Sprintf("gitry - %s [󰘬 %s/%s]", Version, repoName, branch)
	}
	return fmt.Sprintf("gitry - %s", Version)
}
