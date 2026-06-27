package treeflow

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"gitry/internal/ui/common"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	diffTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(common.ColorFrostBlue).
			PaddingLeft(1)

	diffBodyStyle = lipgloss.NewStyle().
			Foreground(common.ColorSnowDark)

	diffAddLineStyle = lipgloss.NewStyle().
				Foreground(common.ColorGreen).
				Background(common.ColorBgDiffAdd)

	diffDelLineStyle = lipgloss.NewStyle().
				Foreground(common.ColorRed).
				Background(common.ColorBgDiffDelete)
)

type AddFilesDiffModel struct {
	path     string
	isDir    bool
	viewport viewport.Model
	width    int
	height   int
	detailed bool
	ready    bool
}

type afDiffBackMsg struct{}

func newAddFilesDiffModel(path string, isDir bool, width, height int) AddFilesDiffModel {
	vp := viewport.New(width, height-3)
	vp.YPosition = 3

	m := AddFilesDiffModel{
		path:     path,
		isDir:    isDir,
		viewport: vp,
		width:    width,
		height:   height,
		detailed: false,
		ready:    true,
	}
	m.rebuildContent()
	return m
}

func (m AddFilesDiffModel) Init() tea.Cmd {
	return nil
}

func (m AddFilesDiffModel) Update(msg tea.Msg) (AddFilesDiffModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 3
		m.rebuildContent()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc", "q":
			return m, func() tea.Msg { return afDiffBackMsg{} }
		case "o":
			m.detailed = !m.detailed
			m.rebuildContent()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m AddFilesDiffModel) View() string {
	toggleHint := "o for entire file"
	if m.detailed {
		toggleHint = "o for diff only"
	}

	label := m.path
	if m.isDir {
		label = m.path + "/"
	}

	title := diffTitleStyle.Render(
		fmt.Sprintf("viewing: '%s'", label),
	)

	shortcuts := []common.Shortcut{
		{Key: "esc/q", Desc: "back"},
		{Key: "j/k", Desc: "scroll"},
		{Key: "o", Desc: toggleHint},
	}
	footer := "\n  " + common.RenderShortcuts(shortcuts) + "\n"

	return title + "\n\n" + m.viewport.View() + footer
}

func (m *AddFilesDiffModel) rebuildContent() {
	var content string

	if m.detailed && !m.isDir {
		// read file from disk
		data, err := os.ReadFile(m.path)
		if err != nil {
			content = diffBodyStyle.Render("could not read file or file is deleted: " + err.Error())
		} else {
			content = diffBodyStyle.Render(string(data))
		}
	} else if m.detailed && m.isDir {
		content = diffBodyStyle.Render("cannot show entire contents of a directory, press 'o' to view diff instead.")
	} else {
		cmd := exec.Command("git", "diff", "HEAD", "--", m.path)
		out, _ := cmd.Output()
		diffStr := string(out)

		if diffStr == "" {
			// maybe it's not tracked
			cmdUntracked := exec.Command("git", "diff", "--no-index", "/dev/null", m.path)
			outUntracked, _ := cmdUntracked.Output()
			diffStr = string(outUntracked)
		}

		if diffStr == "" {
			content = diffBodyStyle.Render("no diff available (file might be empty, binary, or unchanged)")
		} else {
			content = m.renderDiffLines(diffStr, m.width)
		}
	}

	m.viewport.SetContent(content)
	m.viewport.GotoTop()
}

func (m *AddFilesDiffModel) renderDiffLines(raw string, w int) string {
	var diffLines []string
	lineNum := 0
	lines := strings.Split(raw, "\n")

	for _, l := range lines {
		if strings.HasPrefix(l, "@@") {
			lineNum = parseHunkStart(l)
			continue
		}
		if strings.HasPrefix(l, "+++ ") || strings.HasPrefix(l, "--- ") {
			continue
		}
		if strings.HasPrefix(l, "diff ") || strings.HasPrefix(l, "index ") ||
			strings.HasPrefix(l, "new file") || strings.HasPrefix(l, "deleted file") ||
			strings.HasPrefix(l, "old mode") || strings.HasPrefix(l, "new mode") ||
			strings.HasPrefix(l, "similarity") || strings.HasPrefix(l, "rename") ||
			strings.HasPrefix(l, "copy") || strings.HasPrefix(l, "Binary") {
			continue
		}
		if strings.HasPrefix(l, "+") {
			diffLines = append(diffLines, diffAddLineStyle.Render(fmt.Sprintf("%d + | %s", lineNum, l[1:])))
			lineNum++
		} else if strings.HasPrefix(l, "-") {
			diffLines = append(diffLines, diffDelLineStyle.Render(fmt.Sprintf("%d - | %s", lineNum, l[1:])))
		} else if strings.HasPrefix(l, " ") {
			diffLines = append(diffLines, diffBodyStyle.Render(fmt.Sprintf("%d   | %s", lineNum, l[1:])))
			lineNum++
		} else if l != "" {
			diffLines = append(diffLines, diffBodyStyle.Render(l))
		}
	}

	return strings.Join(diffLines, "\n")
}

func parseHunkStart(hunk string) int {
	idx := strings.Index(hunk, "+")
	if idx < 0 {
		return 0
	}
	rest := hunk[idx+1:]
	comma := strings.IndexAny(rest, ", ")
	if comma > 0 {
		rest = rest[:comma]
	}
	n, _ := strconv.Atoi(rest)
	return n
}
