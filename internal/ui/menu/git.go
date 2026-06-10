package menu

import (
	"fmt"

	"gitty/internal/git"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type GitChoiceMsg struct {
	Choice string
}

var gitTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#a78bfa")).
	PaddingLeft(2)

type GitModel struct {
	list   list.Model
	choice string
	width  int
	height int
}

// the main git menu
func NewGit(width, height int) GitModel {
	items := []list.Item{
		item{title: "All Git Tools", desc: "merge, rebase, reset, restore, fetch, pull, status and more"},
		item{title: "Commit", desc: "stage, write and push a commit"},
		item{title: "Project Tree", desc: "view and manage tracked files"},
		item{title: "Commit History", desc: "browse the commit log with a nice graph"},
		item{title: "Change Branch", desc: "switch to a different branch"},
		item{title: "About gitty", desc: "info about gitty"},
		item{title: "Quit", desc: "exit gitty :("},
	}

	branch := git.CurrentBranch()
	repoName := git.RepoName()
	l := list.New(items, list.NewDefaultDelegate(), width, height)
	l.Title = fmt.Sprintf("gitty - v0.1.0 (unstable; not yet released) [ %s/%s]", repoName, branch)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	l.Styles.Title = gitTitleStyle

	return GitModel{list: l, width: width, height: height}
}

// no init command needed
func (m GitModel) Init() tea.Cmd {
	return nil
}

// handle input and window resizes
func (m GitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					return GitChoiceMsg{Choice: selected.title}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// render the list
func (m GitModel) View() string {
	return m.list.View()
}
