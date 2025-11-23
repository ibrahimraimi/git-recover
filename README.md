# Git Recover

Git Recover is a simple CLI tool written in Go that helps you recover accidentally deleted git branches. It scans for dangling commits and reflog entries to find lost work, presents them in an interactive TUI, and allows you to restore them to a new branch.>

> Click to open video demo
> [![Git Recover Demo](https://i.pinimg.com/1200x/a2/e2/9f/a2e29fc13256aecd5f4b700fc53c3e7b.jpg)](https://res.cloudinary.com/dajnn3jbw/video/upload/v1763867184/git-recover-demo-II_ucznxo.mp4)

## Installation

```bash
go install github.com/ibrahimraimi/git-recover@latest
```

Or build from source:

```bash
git clone https://github.com/ibrahimraimi/git-recover.git
cd git-recover
go build -o git-recover
```

## Features

- **Recover Lost Commits**: Finds dangling commits and reflog entries.
- **Interactive TUI**: Browse commits with a user-friendly interface.
- **Commit Preview**: View commit details and diffs in a split-pane view before recovering.
- **Cross-Platform**: Works on Linux, macOS, and Windows.

## Usage

Navigate to your git repository and run the tool:

```bash
git-recover
```

### Controls

- **Up / k**: Move cursor up
- **Down / j**: Move cursor down
- **Enter**: Select commit to recover
- **Esc**: Cancel selection / Quit
- **q / Ctrl+c**: Quit

When you select a commit, you will be prompted to enter a name for the new branch. Press **Enter** to confirm and create the branch.

## How it works

The tool uses `git fsck --lost-found` to find dangling commits and `git reflog` to find recent HEAD movements. It aggregates these commits and displays them in a list. When you choose to recover a commit, it simply runs `git branch <new-branch-name> <commit-hash>`.

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to submit pull requests and report issues.

## License

MIT
