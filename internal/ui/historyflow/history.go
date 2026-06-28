package historyflow

import (
	"fmt"
	"strings"

	"github.com/rokybeast/zengit/internal/git"
	"github.com/rokybeast/zengit/internal/ui/common"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BackMsg struct{}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(common.ColorFrostBlue). // nord frost blue
			PaddingLeft(2)

	nodeStyle = lipgloss.NewStyle().
			Foreground(common.ColorFrostBlue). // nord frost blue
			Bold(true)

	graphLineStyle = lipgloss.NewStyle().
			Foreground(common.ColorMutedGray) // nord muted gray

	msgStyle = lipgloss.NewStyle().
			Foreground(common.ColorSnow) // nord snow

	hashStyle = lipgloss.NewStyle().
			Foreground(common.ColorFrostLightBlue) // nord frost

	cursorStyle = lipgloss.NewStyle().
			Foreground(common.ColorFrostBlue). // nord frost blue
			Bold(true)

	hintStyle = lipgloss.NewStyle().
			Foreground(common.ColorMutedGray). // nord muted gray
			PaddingLeft(2)

	headerBranchStyle = lipgloss.NewStyle().
				Foreground(common.ColorGreen). // nord green
				Bold(true)
)

type Model struct {
	rows      []git.GraphRow
	cursor    int
	width     int
	height    int
	sha       string
	showDiff  bool
	diffModel DiffModel
}

func New(width, height int) Model {
	rows, _ := git.BuildGraph()
	sha := git.LatestShortSHA()

	m := Model{
		rows:   rows,
		cursor: 0,
		width:  width,
		height: height,
		sha:    sha,
	}

	for m.cursor < len(m.rows)-1 && m.rows[m.cursor].IsRoute {
		m.cursor++
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// IF we in the diff view, delegate everything there
	if m.showDiff {
		switch msg.(type) {
		case diffBackMsg:
			m.showDiff = false
			return m, nil
		}
		var cmd tea.Cmd
		m.diffModel, cmd = m.diffModel.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc", "q":
			return m, func() tea.Msg { return BackMsg{} }
		case "enter":
			// open diff for selected commit
			if m.cursor >= 0 && m.cursor < len(m.rows) && !m.rows[m.cursor].IsRoute {
				commit := m.rows[m.cursor].Commit
				m.diffModel = newDiffModel(commit.Hash, commit.Message, m.width, m.height)
				m.showDiff = true
				return m, nil
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				for m.cursor > 0 && m.rows[m.cursor].IsRoute {
					m.cursor--
				}
			}
		case "down", "j":
			if m.cursor < len(m.rows)-1 {
				m.cursor++
				for m.cursor < len(m.rows)-1 && m.rows[m.cursor].IsRoute {
					m.cursor++
				}
			}
		case "g":
			m.cursor = 0
			for m.cursor < len(m.rows)-1 && m.rows[m.cursor].IsRoute {
				m.cursor++
			}
		case "G":
			m.cursor = len(m.rows) - 1
			for m.cursor > 0 && m.rows[m.cursor].IsRoute {
				m.cursor--
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	// delegate to diff view
	if m.showDiff {
		return m.diffModel.View()
	}

	var view strings.Builder

	repoName := git.RepoName()
	branch := git.CurrentBranch()
	title := titleStyle.Render(fmt.Sprintf("commit graph (%s)", m.sha))
	view.WriteString("\n" + title + "\n\n")

	header := fmt.Sprintf("  %s  %s",
		nodeStyle.Render("󰘬"), // nf-md-source_branch
		headerBranchStyle.Render(fmt.Sprintf("%s/%s", repoName, branch)),
	)
	view.WriteString(header + "\n")
	view.WriteString(graphLineStyle.Render("  │") + "\n")

	if len(m.rows) == 0 {
		view.WriteString(hintStyle.Render("  no commits found.") + "\n")

		shortcuts := []common.Shortcut{
			{Key: "esc/q", Desc: "back"},
		}
		view.WriteString("\n  " + common.RenderShortcuts(shortcuts) + "\n")
		return view.String()
	}

	maxVisible := m.height - 7
	if maxVisible < 1 {
		maxVisible = 1
	}

	start := m.cursor - maxVisible/2
	if start < 0 {
		start = 0
	}
	end := start + maxVisible
	if end > len(m.rows) {
		end = len(m.rows)
		start = end - maxVisible
		if start < 0 {
			start = 0
		}
	}

	for i := start; i < end; i++ {
		row := m.rows[i]
		rendered := renderRow(row, i == m.cursor)
		view.WriteString(rendered + "\n")
	}

	shortcuts := []common.Shortcut{
		{Key: "esc/q", Desc: "back"},
		{Key: "enter", Desc: "details"},
	}

	if len(m.rows) > maxVisible {
		pos := fmt.Sprintf("[%d/%d]", m.cursor+1, len(m.rows))
		shortcuts = append(shortcuts, common.Shortcut{Key: "pos", Desc: pos})
	}

	view.WriteString("\n  " + common.RenderShortcuts(shortcuts) + "\n")

	return view.String()
}

func renderRow(row git.GraphRow, selected bool) string {
	styledGraph := styleGraphChars(row.Prefix)

	if !row.IsRoute {
		node := nodeStyle.Render("") // nf-dev-git-comitt
		styledGraph = strings.Replace(styledGraph, "*", node, 1)
	}

	var styledText string
	if !row.IsRoute {
		var msgStr, hashStr string
		if selected {
			hashStr = cursorStyle.Render(fmt.Sprintf("[%s]", row.Commit.Hash[:7]))
			msgStr = cursorStyle.Render(row.Commit.Message)
		} else {
			hashStr = hashStyle.Render(fmt.Sprintf("[%s]", row.Commit.Hash[:7]))
			msgStr = msgStyle.Render(row.Commit.Message)
		}
		styledText = fmt.Sprintf("%s %s", msgStr, hashStr)
	}

	full := styledGraph + styledText

	if selected && !row.IsRoute {
		prefix := cursorStyle.Render("> ")
		return prefix + full
	}
	return "  " + full
}

// apply nord colors to graph characters
func styleGraphChars(s string) string {
	var result strings.Builder
	runes := []rune(s)

	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		switch ch {
		case '│', '├', '─', '╮', '╯', '╰', '╭', '┼':
			result.WriteString(graphLineStyle.Render(string(ch)))
		case '*':
			result.WriteRune('*') // keep * for replacement later
		case ' ':
			result.WriteRune(' ')
		default:
			result.WriteString(graphLineStyle.Render(string(ch)))
		}
	}

	return result.String()
}
