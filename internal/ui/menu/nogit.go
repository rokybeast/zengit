package menu

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gitry/internal/ui/common"
)

// menu entry ids
const (
	IDInitRepo   = "init_repo"
	IDNavigate   = "navigate"
	IDAbout      = "about"
	IDQuit       = "quit"
	IDAddFiles   = "add_files"
	IDCommit     = "commit"
	IDPush       = "push"
	IDTree       = "tree"
	IDHistory    = "history"
	IDOtherTools = "other_tools"
)

// lipgloss to style the ui
var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(common.ColorFrostBlue). // nord frost blue
		PaddingLeft(2)
)

type item struct {
	id    string
	title string
	desc  string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// id accessor so model.go can read it
func (i item) ID() string { return i.id }

type ChoiceMsg struct {
	ID string
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
		item{id: IDInitRepo, title: "󰳏 Initialize a Git Repository", desc: "set up a new repo with the base files (readme.md, license, .gitignore)"},
		item{id: IDNavigate, title: "󱣱 Navigate to a Git Repository", desc: "browse to an existing repo"},
		item{id: IDAbout, title: "󰋼 About gitry", desc: "info about gitry"},
		item{id: IDQuit, title: "󰈆 Quit", desc: "exit gitry :("},
	}

	l := list.New(items, nordListDelegate(), 0, 0)
	l.Title = "gitry - v0.1.0 (unstable; not yet released)"
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
				m.choice = selected.id
				return m, func() tea.Msg {
					return ChoiceMsg{ID: selected.id}
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
