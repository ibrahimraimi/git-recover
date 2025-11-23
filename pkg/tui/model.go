package tui

import (
	"fmt"
	"git-recover/pkg/git"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type State int

const (
	StateBrowsing State = iota
	StateInputBranch
	StateRecovering
	StateSuccess
	StateError
)

type Model struct {
	commits   []git.Commit
	cursor    int
	state     State
	textInput textinput.Model
	err       error
	msg       string
}

func NewModel(commits []git.Commit) Model {
	ti := textinput.New()
	ti.Placeholder = "New branch name"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return Model{
		commits:   commits,
		state:     StateBrowsing,
		textInput: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state != StateInputBranch {
				return m, tea.Quit
			}
		case "esc":
			if m.state == StateInputBranch {
				m.state = StateBrowsing
				return m, nil
			}
			return m, tea.Quit
		}

		switch m.state {
		case StateBrowsing:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.commits)-1 {
					m.cursor++
				}
			case "enter":
				m.state = StateInputBranch
				m.textInput.SetValue("recovered-" + m.commits[m.cursor].Hash[:7])
			}

		case StateInputBranch:
			switch msg.String() {
			case "enter":
				m.state = StateRecovering
				return m, func() tea.Msg {
					err := git.RecoverBranch(m.commits[m.cursor].Hash, m.textInput.Value())
					if err != nil {
						return errMsg(err)
					}
					return successMsg("Branch recovered successfully!")
				}
			}
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd

		case StateSuccess, StateError:
			return m, tea.Quit
		}

	case errMsg:
		m.err = msg
		m.state = StateError

	case successMsg:
		m.msg = string(msg)
		m.state = StateSuccess
	}

	return m, nil
}

type errMsg error
type successMsg string

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4")).Padding(0, 1)
	itemStyle     = lipgloss.NewStyle().PaddingLeft(2)
	selectedStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

func (m Model) View() string {
	if m.state == StateError {
		return fmt.Sprintf("Error: %v\nPress q to quit.", m.err)
	}
	if m.state == StateSuccess {
		return fmt.Sprintf("%s\nPress q to quit.", m.msg)
	}

	s := titleStyle.Render("Git Recover") + "\n\n"

	if m.state == StateBrowsing {
		for i, choice := range m.commits {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
				s += selectedStyle.Render(fmt.Sprintf("%s %s - %s (%s)", cursor, choice.Hash[:7], choice.Message, choice.Date)) + "\n"
			} else {
				s += itemStyle.Render(fmt.Sprintf("%s %s - %s (%s)", cursor, choice.Hash[:7], choice.Message, choice.Date)) + "\n"
			}
		}
		s += "\nUse j/k to navigate, enter to recover, q to quit.\n"
	} else if m.state == StateInputBranch {
		s += fmt.Sprintf("Recovering commit %s\n\n", m.commits[m.cursor].Hash[:7])
		s += "Enter new branch name:\n"
		s += m.textInput.View()
		s += "\n\n(esc to cancel, enter to confirm)\n"
	} else if m.state == StateRecovering {
		s += "Recovering...\n"
	}

	return s
}
