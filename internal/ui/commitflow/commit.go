package commitflow

import (
	"fmt"
	"os/exec"
	"strings"

	"gitty/internal/git"
	"gitty/internal/ui/common"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// commit wizard states
type step int

const (
	stepChooseType step = iota
	stepCustomPrefix
	stepScope
	stepMessage
	stepDone
)

// sent back to root model when user presses esc/backspace on done screen
type BackMsg struct{}

// nord-themed styles for the commit wizard
var (
	commitTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#88c0d0")). // nord frost blue
				PaddingLeft(2)

	commitHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4c566a")). // nord muted gray
			PaddingLeft(2).
			MarginTop(1)

	commitSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#a3be8c")). // nord green
				Bold(true).
				PaddingLeft(2)

	commitDetailStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#d8dee9")). // nord snow
				PaddingLeft(2)

	commitErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#bf616a")). // nord red
				Bold(true).
				PaddingLeft(2)

	commitDescStyle = lipgloss.NewStyle().
			Italic(true)

	commitInputLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#81a1c1")). // nord frost
				Bold(true).
				PaddingLeft(2).
				MarginBottom(1)
)

// nord-themed list delegate
func nordListDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.NormalTitle = d.Styles.NormalTitle.Foreground(lipgloss.Color("#eceff4"))
	d.Styles.NormalDesc = d.Styles.NormalDesc.Foreground(lipgloss.Color("#4c566a"))
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#88c0d0")).
		BorderLeftForeground(lipgloss.Color("#88c0d0"))
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#81a1c1")).
		BorderLeftForeground(lipgloss.Color("#88c0d0"))
	d.Styles.DimmedTitle = d.Styles.DimmedTitle.Foreground(lipgloss.Color("#4c566a"))
	d.Styles.DimmedDesc = d.Styles.DimmedDesc.Foreground(lipgloss.Color("#4c566a"))
	d.Styles.FilterMatch = d.Styles.FilterMatch.Foreground(lipgloss.Color("#a3be8c"))
	return d
}

// list item for commit types
type commitItem struct {
	title string
	desc  string
}

func (i commitItem) Title() string       { return i.title }
func (i commitItem) Description() string { return commitDescStyle.Render(i.desc) }
func (i commitItem) FilterValue() string { return i.title }

type Model struct {
	step       step
	list       list.Model
	input      textinput.Model
	commitType string
	scope      string
	message    string
	shortSha   string
	longSha    string
	showLong   bool
	commitErr  string
	width      int
	height     int
	isCustom   bool
}

// create a fresh commit wizard
func New(width, height int) Model {
	items := []list.Item{
		commitItem{title: "feat", desc: "feature - when you add something new"},
		commitItem{title: "fix", desc: "fix - its pretty easy to understand, its a fix"},
		commitItem{title: "refactor", desc: "refactor - if you update your code in a way, that changes its location or even update the format, its a refactor"},
		commitItem{title: "docs", desc: "documentation - any update to docfiles, such as mdfiles, its a documentation update"},
		commitItem{title: "chore", desc: "chore - if its a task that is boring, then it is a chore (you hate it)"},
		commitItem{title: "pkg", desc: "package - which means to update/download/remove packages from lockfiles/config files"},
		commitItem{title: "custom commit message", desc: "its your commit, spread your abstractness :D"},
	}

	l := list.New(items, nordListDelegate(), width, height)
	l.Title = "choose a commit type"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	l.Styles.Title = commitTitleStyle

	return Model{
		step:   stepChooseType,
		list:   l,
		width:  width,
		height: height,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

// update handles all the state transitions for the commit wizard
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.step == stepChooseType {
			m.list.SetSize(msg.Width, msg.Height)
		}
		return m, nil

	case tea.KeyMsg:
		// global escape routes
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.step == stepDone {
				return m, func() tea.Msg { return BackMsg{} }
			}
			return m, func() tea.Msg { return BackMsg{} }
		}

		// step-specific key handling
		switch m.step {
		case stepChooseType:
			return m.updateChooseType(msg)
		case stepCustomPrefix:
			return m.updateCustomPrefix(msg)
		case stepScope:
			return m.updateScope(msg)
		case stepMessage:
			return m.updateMessage(msg)
		case stepDone:
			return m.updateDone(msg)
		}
	}

	// pass remaining messages to the active component
	switch m.step {
	case stepChooseType:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	case stepCustomPrefix, stepScope, stepMessage:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handle commit type selection
func (m Model) updateChooseType(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		selected, ok := m.list.SelectedItem().(commitItem)
		if !ok {
			return m, nil
		}

		if selected.title == "custom commit message" {
			// skip all template stuff and go straight to message input
			m.isCustom = true
			m.step = stepMessage
			m.input = newTextInput("write your full commit message...")
			return m, m.input.Focus()
		}

		// standard type selected, skip to scope
		m.commitType = selected.title
		m.step = stepScope
		m.input = newTextInput("e.g., ui, git, menu...")
		return m, m.input.Focus()
	}

	// let the list handle navigation keys
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// handle custom prefix text input
func (m Model) updateCustomPrefix(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		val := strings.TrimSpace(m.input.Value())
		if val == "" {
			return m, nil
		}
		m.commitType = val
		m.step = stepScope
		m.input = newTextInput("e.g., ui, git, menu...")
		return m, m.input.Focus()
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// handle scope text input
func (m Model) updateScope(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.scope = strings.TrimSpace(m.input.Value())
		m.step = stepMessage
		m.input = newTextInput("what did you change?")
		return m, m.input.Focus()
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// handle commit message text input
func (m Model) updateMessage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		val := strings.TrimSpace(m.input.Value())
		if val == "" {
			return m, nil
		}
		m.message = val
		m.step = stepDone

		// build the commit string
		var commitStr string
		if m.isCustom {
			// custom commit: use the message as-is
			commitStr = m.message
		} else if m.scope != "" {
			commitStr = fmt.Sprintf("%s(%s): %s", m.commitType, m.scope, m.message)
		} else {
			commitStr = fmt.Sprintf("%s: %s", m.commitType, m.message)
		}

		// run git commit
		cmd := exec.Command("git", "commit", "-m", commitStr)
		if out, err := cmd.CombinedOutput(); err != nil {
			m.commitErr = strings.TrimSpace(string(out))
			return m, nil
		}

		// grab the short sha
		shortCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
		if out, err := shortCmd.Output(); err == nil {
			m.shortSha = strings.TrimSpace(string(out))
		} else {
			m.shortSha = "??????"
		}

		// grab the long sha
		longCmd := exec.Command("git", "rev-parse", "HEAD")
		if out, err := longCmd.Output(); err == nil {
			m.longSha = strings.TrimSpace(string(out))
		} else {
			m.longSha = "unknown"
		}

		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// handle toggling sha display on the done screen
func (m Model) updateDone(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+o":
		m.showLong = !m.showLong
		return m, nil
	}
	return m, nil
}

// render the current step
func (m Model) View() string {
	switch m.step {
	case stepChooseType:
		return m.list.View()
	case stepCustomPrefix:
		return m.viewTextInput("write your custom prefix", "> e.g., 'hotfix', 'wip', or whatever you want")
	case stepScope:
		return m.viewTextInput("write a scope", "> a scope is a way to tell that you have changed a certain part of the code")
	case stepMessage:
		if m.isCustom {
			return m.viewTextInput("write your commit message", "> its your commit, spread your abstractness :D")
		}
		return m.viewTextInput("write the commit message", "> make sure it's small, and understandable")
	case stepDone:
		return m.viewDone()
	}
	return ""
}

// render a text input with a label and hint
func (m Model) viewTextInput(label, hint string) string {
	labelStr := commitInputLabelStyle.Render(label)
	hintStr := commitHintStyle.Render(hint)
	inputStr := lipgloss.NewStyle().PaddingLeft(2).Render(m.input.View())

	shortcuts := []common.Shortcut{
		{Key: "esc/q", Desc: "back"},
		{Key: "enter", Desc: "confirm"},
	}
	footer := "\n  " + common.RenderShortcuts(shortcuts)

	return lipgloss.JoinVertical(lipgloss.Left, "", labelStr, hintStr, "", inputStr, "", footer)
}

// render the success/error screen after committing
func (m Model) viewDone() string {
	if m.commitErr != "" {
		errMsg := commitErrorStyle.Render("commit failed!")
		detail := commitDetailStyle.Render(m.commitErr)
		
		shortcuts := []common.Shortcut{
			{Key: "esc", Desc: "back"},
		}
		footer := "\n  " + common.RenderShortcuts(shortcuts)
		
		return lipgloss.JoinVertical(lipgloss.Left, "", errMsg, "", detail, "", footer)
	}

	repoName := git.RepoName()
	branch := git.CurrentBranch()

	header := commitSuccessStyle.Render(
		fmt.Sprintf("committed to [\uf126 %s/%s]", repoName, branch),
	)

	sha := m.shortSha
	if m.showLong {
		sha = m.longSha
	}

	shaLine := commitDetailStyle.Render(
		fmt.Sprintf("commit id: %s", sha),
	)

	shortcuts := []common.Shortcut{
		{Key: "esc", Desc: "menu"},
		{Key: "ctrl+o", Desc: "toggle sha length"},
	}
	footer := "\n  " + common.RenderShortcuts(shortcuts)

	return lipgloss.JoinVertical(lipgloss.Left, "", header, "", shaLine, "", footer)
}

// create a styled text input with a placeholder
func newTextInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 120
	ti.Width = 50
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#88c0d0"))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#eceff4"))
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#4c566a"))
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#88c0d0"))
	return ti
}
