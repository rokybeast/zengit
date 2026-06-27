package initflow

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitry/internal/git"
	"gitry/internal/ui/common"
)

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(common.ColorFrostBlue). // nord frost blue
	PaddingLeft(2)

// nord-themed list delegate
func nordListDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.NormalTitle = d.Styles.NormalTitle.Foreground(common.ColorSnow)
	d.Styles.NormalDesc = d.Styles.NormalDesc.Foreground(common.ColorMutedGray)
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(common.ColorFrostBlue).
		BorderLeftForeground(common.ColorFrostBlue)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(common.ColorFrostLightBlue).
		BorderLeftForeground(common.ColorFrostBlue)
	d.Styles.DimmedTitle = d.Styles.DimmedTitle.Foreground(common.ColorMutedGray)
	d.Styles.DimmedDesc = d.Styles.DimmedDesc.Foreground(common.ColorMutedGray)
	d.Styles.FilterMatch = d.Styles.FilterMatch.Foreground(common.ColorGreen)
	return d
}

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
		return "initializing a git repository...\n"
	}
	return m.list.View()
}

// .gitignore list
func (m Model) loadGitIgnoreList() Model {
	items := make([]list.Item, len(git.GitIgnores))
	for i, t := range git.GitIgnores {
		items[i] = item{template: t}
	}
	l := list.New(items, nordListDelegate(), m.width, m.height)
	l.Title = "choose a .gitignore template"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	m.list = l
	return m
}

// license list
func (m Model) loadLicenseList() Model {
	items := make([]list.Item, len(git.Licenses))
	for i, t := range git.Licenses {
		items[i] = item{template: t}
	}
	l := list.New(items, nordListDelegate(), m.width, m.height)
	l.Title = "choose a license (gitry does not warn you about the legality of licenses, for that, contact saul goodman!)"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	m.list = l
	return m
}
