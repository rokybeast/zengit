package menu

import (
	"fmt"

	"gitry/internal/git"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gitry/internal/ui/common"
)

type GitChoiceMsg struct {
	ID string
}

var gitTitleStyle = lipgloss.NewStyle().
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

type GitModel struct {
	list   list.Model
	choice string
	width  int
	height int
}

// the main git menu
func NewGit(width, height int) GitModel {
	items := []list.Item{
		item{id: IDAddFiles, title: "󰝒 Add Files", desc: "stage or unstage files for commit"},
		item{id: IDCommit, title: "󰜘 Commit", desc: "stage and write commits"},
		item{id: IDPush, title: " Push Commits", desc: "push the commits to different remotes"},
		item{id: IDTree, title: "󰙅 Project Tree", desc: "view and manage tracked files"},
		item{id: IDHistory, title: "󰋚 Commit History", desc: "browse the commit log with a nice graph"},
		item{id: IDOtherTools, title: "󱈧 Other Git Tools", desc: "merge, rebase, reset, restore, fetch, pull, status and more..."},
		item{id: IDAbout, title: "󰋼 About gitry", desc: "info about gitry"},
		item{id: IDQuit, title: "󰈆 Quit", desc: "exit gitry :("},
	}

	branch := git.CurrentBranch()
	repoName := git.RepoName()
	l := list.New(items, nordListDelegate(), width, height)
	l.Title = fmt.Sprintf("gitry - v0.3.0 (unstable; not yet released) [󰘬 %s/%s]", repoName, branch)
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
				m.choice = selected.id
				return m, func() tea.Msg {
					return GitChoiceMsg{ID: selected.id}
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
