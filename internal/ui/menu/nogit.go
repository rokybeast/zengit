package menu

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// lipgloss to style the ui
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#38bdf8")). // nord blue (we ALL *somewhat* love nord)
			PaddingLeft(2)

	itemStyle         = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#94a3b8")) // nord gray
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#e2e8f0")) // nord bright white
)

type item struct {
	title string
	desc  string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type ChoiceMsg struct {
	Choice string
}

type NoGitModel struct {
	list   list.Model
	choice string
	width  int
	height int
}

// the no git menu (its useful)
func NewNoGit() NoGitModel {
	items := []list.Item{
		item{title: "Initialize a Git Repository", desc: "set up a new repo with the base files (README.md, LICENSE, .gitignore)"},
		item{title: "Navigate to a Git Repository", desc: "browse to an existing repo"},
		item{title: "About gitty", desc: "info about gitty"},
		item{title: "Quit", desc: "exit gitty :("},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "gitty - v0.1.0 (unstable; not yet released)"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	l.Styles.Title = titleStyle

	return NoGitModel{list: l}
}

// no init. command
func (m NoGitModel) Init() tea.Cmd {
	return nil
}

// survive the input and the dangerous window resize (its scary for a tui application)
func (m NoGitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			selected, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = selected.title
				return m, func() tea.Msg {
					return ChoiceMsg{Choice: selected.title}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// render list
func (m NoGitModel) View() string {
	return m.list.View()
}
