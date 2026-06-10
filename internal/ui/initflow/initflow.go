package initflow

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"gitty/internal/git"
)

type step int

const (
	stepGitIgnore step = iota
	stepLicense
	stepDone
)

type DoneMsg struct {
	GitIgnore git.Template
	License   git.Template
}

type item struct {
	template git.Template
}

func (i item) Title() string       { return i.template.Name }
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.template.Name }

type Model struct {
	step      step
	list      list.Model
	gitIgnore git.Template
	license   git.Template
	width     int
	height    int
}

func New(width, height int) Model {
	m := Model{step: stepGitIgnore, width: width, height: height}
	m = m.loadGitIgnoreList()
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

// update workflow state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selected, ok := m.list.SelectedItem().(item)
			if ok {
				if m.step == stepGitIgnore {
					m.gitIgnore = selected.template
					m.step = stepLicense
					m = m.loadLicenseList()
					return m, nil
				} else if m.step == stepLicense {
					m.license = selected.template
					m.step = stepDone
					return m, func() tea.Msg {
						return DoneMsg{GitIgnore: m.gitIgnore, License: m.license}
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.step == stepDone {
		return "Initializing a Git Repository...\n"
	}
	return m.list.View()
}

// .gitignore list
func (m Model) loadGitIgnoreList() Model {
	items := make([]list.Item, len(git.GitIgnores))
	for i, t := range git.GitIgnores {
		items[i] = item{template: t}
	}
	l := list.New(items, list.NewDefaultDelegate(), m.width, m.height)
	l.Title = "Choose a .gitignore template"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	m.list = l
	return m
}

// license list
func (m Model) loadLicenseList() Model {
	items := make([]list.Item, len(git.Licenses))
	for i, t := range git.Licenses {
		items[i] = item{template: t}
	}
	l := list.New(items, list.NewDefaultDelegate(), m.width, m.height)
	l.Title = "Choose a License (gitty does not warn you about the legality of Licenses, for that, contact Saul Goodman!)"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	m.list = l
	return m
}
