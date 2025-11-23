package tui

import (
	"fmt"
	"git-recover/pkg/git"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
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
	viewport  viewport.Model
	err       error
	msg       string
	ready     bool
	width     int
	height    int
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
	return tea.Batch(textinput.Blink, m.updatePreview())
}

func (m Model) updatePreview() tea.Cmd {
	return func() tea.Msg {
		if len(m.commits) == 0 {
			return nil
		}
		content, err := git.GetCommitShow(m.commits[m.cursor].Hash)
		if err != nil {
			return errMsg(err)
		}
		return previewMsg(content)
	}
}

type previewMsg string

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.viewport = viewport.New(msg.Width/2, msg.Height-5)
			m.viewport.HighPerformanceRendering = false
			m.ready = true
		} else {
			m.viewport.Width = msg.Width / 2
			m.viewport.Height = msg.Height - 5
		}

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
					cmds = append(cmds, m.updatePreview())
				}
			case "down", "j":
				if m.cursor < len(m.commits)-1 {
					m.cursor++
					cmds = append(cmds, m.updatePreview())
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
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)

		case StateSuccess, StateError:
			return m, tea.Quit
		}

	case previewMsg:
		m.viewport.SetContent(string(msg))

	case errMsg:
		m.err = msg
		m.state = StateError

	case successMsg:
		m.msg = string(msg)
		m.state = StateSuccess
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

type errMsg error
type successMsg string

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4")).Padding(0, 1)
	itemStyle     = lipgloss.NewStyle().PaddingLeft(2)
	selectedStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	viewStyle     = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("69"))
)

func (m Model) View() string {
	if m.state == StateError {
		return fmt.Sprintf("Error: %v\nPress q to quit.", m.err)
	}
	if m.state == StateSuccess {
		return fmt.Sprintf("%s\nPress q to quit.", m.msg)
	}
	if !m.ready {
		return "Initializing..."
	}

	header := titleStyle.Render("Git Recover") + "\n\n"

	// List View
	var listContent string
	if m.state == StateBrowsing || m.state == StateInputBranch {
		for i, choice := range m.commits {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
				listContent += selectedStyle.Render(fmt.Sprintf("%s %s - %s (%s)", cursor, choice.Hash[:7], choice.Message, choice.Date)) + "\n"
			} else {
				listContent += itemStyle.Render(fmt.Sprintf("%s %s - %s (%s)", cursor, choice.Hash[:7], choice.Message, choice.Date)) + "\n"
			}
		}
		listContent += "\nUse j/k to navigate, enter to recover, q to quit.\n"
	}

	if m.state == StateInputBranch {
		listContent += fmt.Sprintf("\nRecovering commit %s\n", m.commits[m.cursor].Hash[:7])
		listContent += "Enter new branch name:\n"
		listContent += m.textInput.View()
		listContent += "\n\n(esc to cancel, enter to confirm)\n"
	} else if m.state == StateRecovering {
		listContent += "Recovering...\n"
	}

	// Split View
	leftView := lipgloss.NewStyle().Width(m.width / 2).Render(listContent)
	rightView := viewStyle.Width(m.width/2 - 2).Height(m.height - 5).Render(m.viewport.View())

	return lipgloss.JoinVertical(lipgloss.Left, header, lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView))
}
