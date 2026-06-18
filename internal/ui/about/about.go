package about

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BackMsg struct{}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#88c0d0")). // nord frost blue
			MarginBottom(1)

	bodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d8dee9")). // nord snow
			PaddingLeft(2)

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4c566a")). // nord muted gray
			PaddingLeft(2).
			MarginTop(1)
)

type Model struct{}

func New() Model { return Model{} }

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "backspace":
			return m, func() tea.Msg { return BackMsg{} }
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	title := titleStyle.Render("about gitty")
	body := bodyStyle.Render(
		"gitty is a feature-rich, aesthetically good looking and mainly, a minimal TUI tool.\n" +
			"built with golang, and with 󰋑 for  \n\n" +
			"version: 0.3.0 alpha\n" + // yes, static version for now..
			"github repository (want to contribute?): github.com/rokybeast/gitty",
	)
	hint := hintStyle.Render("press esc to go back")

	return lipgloss.JoinVertical(lipgloss.Left, title, body, hint)
}
