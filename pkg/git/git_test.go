package git

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestGetDanglingCommits(t *testing.T) {
	// Setup temporary git repo
	dir, err := os.MkdirTemp("", "git-recover-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	// Create a commit
	execCommand(t, dir, "git", "checkout", "-b", "main")
	createCommit(t, dir, "initial commit")

	// Create a branch and a commit, then delete the branch
	execCommand(t, dir, "git", "checkout", "-b", "feature")
	createCommit(t, dir, "feature commit")
	hash := getHeadHash(t, dir)
	execCommand(t, dir, "git", "checkout", "main")
	execCommand(t, dir, "git", "branch", "-D", "feature")

	// Change working directory to the temp repo for the test
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(cwd)

	// Test GetDanglingCommits
	commits, err := GetDanglingCommits()
	if err != nil {
		t.Fatalf("GetDanglingCommits failed: %v", err)
	}

	found := false
	for _, c := range commits {
		if c.Hash == hash {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find dangling commit %s, but didn't. Found: %v", hash, commits)
	}
}

func createCommit(t *testing.T, dir, msg string) {
	execCommand(t, dir, "git", "commit", "--allow-empty", "-m", msg)
}

func execCommand(t *testing.T, dir, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Command %s %v failed: %v\nOutput: %s", name, args, err, out)
	}
}

func getHeadHash(t *testing.T, dir string) string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}
