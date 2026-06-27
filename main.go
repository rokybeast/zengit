package main

import (
	"fmt"
	"os"

	"gitry/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(ui.New(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[error]: gitry error - %v\n", err)
		os.Exit(1)
	}
}
