package nav

import (
	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitty/internal/git"
)

type PickedMsg struct {
	Path string
}

type BackMsg struct{}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#c4b5fd")).
			PaddingLeft(2).
			MarginBottom(1)

	errStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f87171")).
			PaddingLeft(2)
)

type Model struct {
	picker filepicker.Model
	err    string
}

// make a file picker (from ~ [/home/ur-name])
func New() Model {
	fp := filepicker.New()
	fp.DirAllowed = true
	fp.FileAllowed = false
	fp.ShowHidden = true

	return Model{picker: fp}
}

func (m Model) Init() tea.Cmd {
	return m.picker.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg { return BackMsg{} }
		}
	}

	var cmd tea.Cmd
	m.picker, cmd = m.picker.Update(msg)

	// check if dir was selected
	if didSelect, path := m.picker.DidSelectFile(msg); didSelect {
		if git.IsRepo(path) {
			m.err = ""
			return m, func() tea.Msg { return PickedMsg{Path: path} }
		}
		m.err = "This is not a Git repository: " + path
	}

	return m, cmd
}

func (m Model) View() string {
	title := titleStyle.Render("Navigate to a Git repository")
	view := m.picker.View()

	if m.err != "" {
		view += "\n" + errStyle.Render(m.err)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, view)
}
