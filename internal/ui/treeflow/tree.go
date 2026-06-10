package treeflow

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"gitty/internal/git"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BackMsg struct{}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#a78bfa")).
			PaddingLeft(4)

	stagedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a3be8c")) // nord green

	unstagedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#bf616a")) // nord red

	untrackedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4c566a")) // nord muted gray

	dirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#81a1c1")). // nord frost
			Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#88c0d0")). // nord lighter blue
			Bold(true)

	defaultStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#eceff4")) // nord snow
)

type treeNode struct {
	path   string
	name   string
	isDir  bool
	prefix string
	status string
}

type Model struct {
	nodes      []treeNode
	cursor     int
	width      int
	height     int
	latestSHA  string
	isAddFiles bool
}

func New(width, height int, isAddFiles bool) Model {
	m := Model{
		width:      width,
		height:     height,
		isAddFiles: isAddFiles,
	}
	m.refresh()
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

// refresh tree data and git statuses
func (m *Model) refresh() {
	statuses := getGitStatus()
	m.nodes = buildTree(".", "", statuses)
	m.latestSHA = git.LatestShortSHA()

	if m.cursor >= len(m.nodes) {
		m.cursor = len(m.nodes) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc", "enter":
			return m, func() tea.Msg { return BackMsg{} }
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.nodes)-1 {
				m.cursor++
			}
		case " ", "a":
			if len(m.nodes) > 0 {
				node := m.nodes[m.cursor]
				toggleStaging(node)
				m.refresh()
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	var view strings.Builder
	if m.isAddFiles {
		view.WriteString(titleStyle.Render("add files (space/a to stage/unstage, enter/esc to go back)") + "\n\n")
	} else {
		view.WriteString(titleStyle.Render(fmt.Sprintf("project tree (%s)", m.latestSHA)) + "\n\n")
	}

	if len(m.nodes) == 0 {
		return view.String() + "  no files found.\n"
	}

	maxItems := m.height - 4
	if maxItems < 1 {
		maxItems = 1
	}

	start := m.cursor - maxItems/2
	if start < 0 {
		start = 0
	}
	end := start + maxItems
	if end > len(m.nodes) {
		end = len(m.nodes)
		start = end - maxItems
		if start < 0 {
			start = 0
		}
	}

	for i := start; i < end; i++ {
		node := m.nodes[i]

		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		colorStyle := defaultStyle
		icon := "󰈔"
		if node.isDir {
			colorStyle = dirStyle
			icon = ""
		} else if node.status == "??" {
			colorStyle = untrackedStyle
		} else if isStaged(node.status) {
			colorStyle = stagedStyle
		} else if node.status != "" {
			colorStyle = unstagedStyle
		}

		statusBlock := ""
		if node.status != "" {
			statusBlock = fmt.Sprintf(" [%s]", node.status)
		}

		prefixPart := fmt.Sprintf("%s%s ", cursor, node.prefix)
		contentPart := fmt.Sprintf("%s %s%s", icon, node.name, statusBlock)

		if i == m.cursor {
			view.WriteString(cursorStyle.Render(prefixPart+contentPart) + "\n")
		} else {
			view.WriteString(prefixPart + colorStyle.Render(contentPart) + "\n")
		}
	}

	return view.String()
}

// fetch git status for the repo
func getGitStatus() map[string]string {
	m := make(map[string]string)
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return m
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}
		status := line[:2]
		parts := strings.SplitN(line[3:], " -> ", 2)
		path := parts[len(parts)-1]
		path = strings.Trim(path, "\"")
		m[path] = status
	}
	return m
}

// find out if a file is staged
func isStaged(status string) bool {
	if len(status) < 2 {
		return false
	}
	c := status[0]
	return c != ' ' && c != '?'
}

// toggle git staging status
func toggleStaging(node treeNode) {
	if isStaged(node.status) {
		cmd := exec.Command("git", "reset", "HEAD", "--", node.path)
		_ = cmd.Run()
	} else {
		cmd := exec.Command("git", "add", "--", node.path)
		_ = cmd.Run()
	}
}

// build tree nodes repeatedly
func buildTree(root string, prefix string, statuses map[string]string) []treeNode {
	var nodes []treeNode
	entries, err := os.ReadDir(root)
	if err != nil {
		return nodes
	}

	var filtered []os.DirEntry
	for _, e := range entries {
		if e.Name() == ".git" {
			continue
		}
		filtered = append(filtered, e)
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].IsDir() != filtered[j].IsDir() {
			return filtered[i].IsDir()
		}
		return filtered[i].Name() < filtered[j].Name()
	})

	for i, entry := range filtered {
		isLast := i == len(filtered)-1

		currPrefix := prefix
		if isLast {
			currPrefix += "└─ "
		} else {
			currPrefix += "├─ "
		}

		fullPath := filepath.Join(root, entry.Name())
		relPath := filepath.Clean(fullPath)
		gitPath := filepath.ToSlash(relPath)

		nodes = append(nodes, treeNode{
			path:   relPath,
			name:   entry.Name(),
			isDir:  entry.IsDir(),
			prefix: currPrefix,
			status: statuses[gitPath],
		})

		if entry.IsDir() {
			newPrefix := prefix
			if isLast {
				newPrefix += "   "
			} else {
				newPrefix += "│  "
			}
			nodes = append(nodes, buildTree(fullPath, newPrefix, statuses)...)
		}
	}

	return nodes
}
