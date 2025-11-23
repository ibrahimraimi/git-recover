package git

import (
	"fmt"
	"os/exec"
	"strings"
)

type Commit struct {
	Hash    string
	Message string
	Author  string
	Date    string
	Type    string // "dangling" or "reflog"
}

// RunCommand runs a git command and returns the output
func RunCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git command failed: %s, output: %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// GetDanglingCommits finds dangling commits using git fsck
func GetDanglingCommits() ([]Commit, error) {
	out, err := RunCommand("fsck", "--lost-found")
	if err != nil {
		return nil, err
	}

	var commits []Commit
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "dangling commit") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				hash := parts[2]
				details, err := getCommitDetails(hash)
				if err == nil {
					details.Type = "dangling"
					commits = append(commits, details)
				}
			}
		}
	}
	return commits, nil
}

// GetReflogCommits finds commits from reflog
func GetReflogCommits() ([]Commit, error) {
	// %H: commit hash, %cd: commit date, %an: author name, %s: subject
	out, err := RunCommand("reflog", "--format=%H|%cd|%an|%s", "-n", "50")
	if err != nil {
		return nil, err
	}

	var commits []Commit
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) == 4 {
			commits = append(commits, Commit{
				Hash:    parts[0],
				Date:    parts[1],
				Author:  parts[2],
				Message: parts[3],
				Type:    "reflog",
			})
		}
	}
	return commits, nil
}

func getCommitDetails(hash string) (Commit, error) {
	out, err := RunCommand("show", "-s", "--format=%H|%cd|%an|%s", hash)
	if err != nil {
		return Commit{}, err
	}
	parts := strings.Split(strings.TrimSpace(out), "|")
	if len(parts) == 4 {
		return Commit{
			Hash:    parts[0],
			Date:    parts[1],
			Author:  parts[2],
			Message: parts[3],
		}, nil
	}
	return Commit{}, fmt.Errorf("failed to parse commit details")
}

// RecoverBranch creates a new branch pointing to the commit
func RecoverBranch(commitHash, branchName string) error {
	_, err := RunCommand("branch", branchName, commitHash)
	return err
}

// PushBranch pushes the branch to remote
func PushBranch(branchName string) error {
	_, err := RunCommand("push", "-u", "origin", branchName)
	return err
}
