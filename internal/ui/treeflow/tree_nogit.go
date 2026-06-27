package treeflow

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gitry/internal/ui/common"
)

type PickedMsg struct {
	Path string
}

var (
	gitRepoStyle = lipgloss.NewStyle().
			Foreground(common.ColorGreen). // nord green
			Bold(true)

	errStyle = lipgloss.NewStyle().
			Foreground(common.ColorRed). // nord red
			PaddingLeft(4)
)

type noGitTreeNode struct {
	path     string
	name     string
	isDir    bool
	isGit    bool
	repoName string
	branch   string
	prefix   string
	expanded bool
}

type NoGitModel struct {
	nodes        []noGitTreeNode
	cursor       int
	width        int
	height       int
	expandedDirs map[string]bool
	err          string
	cwd          string
}

func NewNoGit(width, height int) NoGitModel {
	cwd, _ := os.UserHomeDir()
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	m := NoGitModel{
		width:        width,
		height:       height,
		expandedDirs: make(map[string]bool),
		cwd:          cwd,
	}
	m.expandedDirs[cwd] = true
	m.refresh()
	return m
}

func (m NoGitModel) Init() tea.Cmd {
	return nil
}

func (m *NoGitModel) refresh() {
	m.nodes = buildNoGitTree(m.cwd, "", m.expandedDirs)
	if m.cursor >= len(m.nodes) {
		m.cursor = len(m.nodes) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m NoGitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg { return BackMsg{} }
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.nodes)-1 {
				m.cursor++
			}
		case "a":
			if len(m.nodes) > 0 {
				node := m.nodes[m.cursor]
				if node.isGit {
					return m, func() tea.Msg { return PickedMsg{Path: node.path} }
				} else {
					m.err = "this is not a git repository: " + node.path
				}
			}
		case " ", "enter":
			if len(m.nodes) > 0 {
				node := m.nodes[m.cursor]
				if node.name == ".." {
					parent := filepath.Dir(m.cwd)
					m.cwd = parent
					m.expandedDirs = make(map[string]bool)
					m.expandedDirs[m.cwd] = true
					m.cursor = 0
					m.refresh()
				} else if node.isDir {
					m.expandedDirs[node.path] = !m.expandedDirs[node.path]
					m.refresh()
				}
			}
		}
	}
	return m, nil
}

func (m NoGitModel) View() string {
	var view strings.Builder
	view.WriteString("\n" + titleStyle.Render("navigate to a git repository (space: toggle folder, a: open repo, esc: go back)") + "\n\n")

	if len(m.nodes) == 0 {
		view.WriteString("  no files found.\n")
	} else {
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
				if node.expanded {
					icon = "󰝰"
				} else {
					icon = "󰉋"
				}
			}

			gitLabel := ""
			if node.isGit {
				gitLabel = gitRepoStyle.Render(fmt.Sprintf(" (%s/%s)", node.repoName, node.branch))
			}

			prefixPart := fmt.Sprintf("%s%s ", cursor, node.prefix)

			if i == m.cursor {
				contentPart := fmt.Sprintf("%s %s%s", icon, node.name, gitLabel)
				view.WriteString(cursorStyle.Render(prefixPart+contentPart) + "\n")
			} else {
				iconAndName := colorStyle.Render(fmt.Sprintf("%s %s", icon, node.name))
				view.WriteString(prefixPart + iconAndName + gitLabel + "\n")
			}
		}
	}

	if m.err != "" {
		view.WriteString("\n" + errStyle.Render(m.err))
	}

	return view.String()
}

func buildNoGitTree(root string, prefix string, expanded map[string]bool) []noGitTreeNode {
	var nodes []noGitTreeNode

	if filepath.Clean(root) != filepath.VolumeName(root)+string(os.PathSeparator) && prefix == "" {
		nodes = append(nodes, noGitTreeNode{
			path:  filepath.Dir(root),
			name:  "..",
			isDir: true,
		})
	}

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
		return strings.ToLower(filtered[i].Name()) < strings.ToLower(filtered[j].Name())
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

		isGit := false
		repoName := ""
		branch := ""

		if entry.IsDir() {
			if isGitRoot(fullPath) {
				isGit = true
				repoName = getRepoNameFromPath(fullPath)
				branch = getBranchFromPath(fullPath)
			}
		}

		nodes = append(nodes, noGitTreeNode{
			path:     fullPath,
			name:     entry.Name(),
			isDir:    entry.IsDir(),
			isGit:    isGit,
			repoName: repoName,
			branch:   branch,
			prefix:   currPrefix,
			expanded: expanded[fullPath],
		})

		if entry.IsDir() && expanded[fullPath] {
			newPrefix := prefix
			if isLast {
				newPrefix += "   "
			} else {
				newPrefix += "│  "
			}
			nodes = append(nodes, buildNoGitTree(fullPath, newPrefix, expanded)...)
		}
	}

	return nodes
}

func isGitRoot(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

func getRepoNameFromPath(path string) string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return filepath.Base(path)
	}
	return filepath.Base(strings.TrimSpace(string(out)))
}

func getBranchFromPath(path string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}
