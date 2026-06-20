package common

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	ShortcutKeyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).
				Bold(true)

	ShortcutDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#525252")) // nord muted gray

	ShortcutSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#525252")).
				Padding(0, 1) // " | " padding
)

type Shortcut struct {
	Key  string
	Desc string
}

// e.g., "esc/q back • space smth • a lol" (im copying the home screen style)
func RenderShortcuts(shortcuts []Shortcut) string {
	var parts []string

	for _, s := range shortcuts {
		key := ShortcutKeyStyle.Render(s.Key)
		desc := ShortcutDescStyle.Render(s.Desc)
		parts = append(parts, key+ShortcutDescStyle.Render()+" "+desc)
	}

	separator := ShortcutSeparatorStyle.Render("•")
	return strings.Join(parts, separator)
}
