package main

import (
	"fmt"
	"git-recover/pkg/git"
	"git-recover/pkg/tui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Check if we are in a git repo
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		fmt.Println("Error: Not a git repository. Please run this command from the root of a git repository.")
		os.Exit(1)
	}

	fmt.Println("Scanning for recoverable commits...")

	dangling, err := git.GetDanglingCommits()
	if err != nil {
		fmt.Printf("Error finding dangling commits: %v\n", err)
	}

	reflog, err := git.GetReflogCommits()
	if err != nil {
		fmt.Printf("Error reading reflog: %v\n", err)
	}

	// Combine and deduplicate
	commitMap := make(map[string]git.Commit)
	for _, c := range dangling {
		commitMap[c.Hash] = c
	}
	for _, c := range reflog {
		if _, exists := commitMap[c.Hash]; !exists {
			commitMap[c.Hash] = c
		}
	}

	var allCommits []git.Commit
	for _, c := range commitMap {
		allCommits = append(allCommits, c)
	}

	if len(allCommits) == 0 {
		fmt.Println("No recoverable commits found.")
		return
	}

	p := tea.NewProgram(tui.NewModel(allCommits))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
