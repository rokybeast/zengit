package about

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BackMsg struct{}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#c4b5fd")).
			MarginBottom(1)

	bodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94a3b8")).
			PaddingLeft(2)

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#64748b")).
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
	title := titleStyle.Render("About gitty")
	body := bodyStyle.Render(
		"gitty is a feature-rich, aesthetically good looking and mainly, a minimal TUI tool.\n" +
			"Built with Golang, and with 󰋑 for  \n\n" +
			"Version: 0.1.0 Alpha\n" + // yes, static version for now..
			"GitHub Repository (Want to contribute?): github.com/rokybeast/gitty",
	)
	hint := hintStyle.Render("Press Esc to Go Back")

	return lipgloss.JoinVertical(lipgloss.Left, title, body, hint)
}
