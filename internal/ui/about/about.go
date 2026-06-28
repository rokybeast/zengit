package about

import (
	"github.com/rokybeast/zengit/internal/ui/common"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BackMsg struct{}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(common.ColorFrostBlue). // nord frost blue
			MarginBottom(1)

	bodyStyle = lipgloss.NewStyle().
			Foreground(common.ColorSnowDark). // nord snow
			PaddingLeft(2)

	hintStyle = lipgloss.NewStyle().
			Foreground(common.ColorMutedGray). // nord muted gray
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
	title := titleStyle.Render("    about zengit")
	body := bodyStyle.Render(
		"  zengit is a feature-rich, aesthetically good looking and mainly, a minimal TUI tool.\n" +
			"  built with golang, and with 󰋑 for  \n\n" +
			"  version: " + common.Version + "\n" +
			"  github repository (want to contribute?): github.com/rokybeast/zengit",
	)

	shortcuts := []common.Shortcut{
		{Key: "esc", Desc: "back"},
	}
	footer := "\n  " + common.RenderShortcuts(shortcuts)

	return lipgloss.JoinVertical(lipgloss.Left, title, body, footer)
}
