package treeflow

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"gitty/internal/ui/common"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	afTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(common.ColorFrostBlue). // nord frost blue
			PaddingLeft(2)

	afStagedStyle = lipgloss.NewStyle().
			Foreground(common.ColorGreen) // nord green

	afUnstagedStyle = lipgloss.NewStyle().
			Foreground(common.ColorRed) // nord red

	afUntrackedStyle = lipgloss.NewStyle().
				Foreground(common.ColorOrange) // nord orange

	afDeletedStyle = lipgloss.NewStyle().
			Foreground(common.ColorRed). // nord red
			Strikethrough(true)

	afCursorStyle = lipgloss.NewStyle().
			Foreground(common.ColorFrostBlue). // nord frost blue
			Bold(true)

	afHintStyle = lipgloss.NewStyle().
			Foreground(common.ColorMutedGray). // nord muted gray
			PaddingLeft(2)

	afDirStyle = lipgloss.NewStyle().
			Foreground(common.ColorFrostLightBlue). // nord frost
			Bold(true)

	afHeaderStyle = lipgloss.NewStyle().
			Foreground(common.ColorSnow). // nord snow
			Bold(true).
			PaddingLeft(2)
)

// a single entry in the flat list (either dir or file)
type dirtyFile struct {
	path      string
	name      string
	status    string
	staged    bool
	deleted   bool
	untracked bool
	dir       string
	isDir     bool
}

// only shows dirty, untracked, and deleted files (grouped by dir)
type AddFilesModel struct {
	allFiles     []dirtyFile
	entries      []dirtyFile
	cursor       int
	width        int
	height       int
	expandedDirs map[string]bool
}

// make a new addfiles model by scanning git status
func NewAddFiles(width, height int) AddFilesModel {
	m := AddFilesModel{
		width:        width,
		height:       height,
		expandedDirs: make(map[string]bool),
	}
	m.refresh()
	return m
}

func (m AddFilesModel) Init() tea.Cmd {
	return nil
}

func (m AddFilesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
		case "a":
			if len(m.entries) > 0 && m.cursor < len(m.entries) {
				entry := m.entries[m.cursor]
				if entry.isDir {
					// toggle all files in this directory
					afToggleDirStaging(m.allFiles, entry.dir)
				} else {
					afToggleStaging(entry)
				}
				m.refresh()
			}
		case " ":
			if len(m.entries) > 0 && m.cursor < len(m.entries) {
				entry := m.entries[m.cursor]
				if entry.isDir {
					m.expandedDirs[entry.dir] = !m.expandedDirs[entry.dir]
					m.refresh()
				}
			}
		case "A":
			// stage all at once (for 'em bulk committers)
			afStageAll()
			m.refresh()
		case "g":
			m.cursor = 0
		case "G":
			if len(m.entries) > 0 {
				m.cursor = len(m.entries) - 1
			}
		}
	}
	return m, nil
}

func (m AddFilesModel) View() string {
	var view strings.Builder

	title := afTitleStyle.Render("add files")
	view.WriteString("\n" + title + "\n\n")

	if len(m.entries) == 0 {
		view.WriteString(afHintStyle.Render("  nothing to stage, working tree is clean.") + "\n")
		shortcuts := []common.Shortcut{
			{Key: "esc/q", Desc: "back"},
		}
		view.WriteString("\n  " + common.RenderShortcuts(shortcuts) + "\n")
		return view.String()
	}

	// count staged vs unstaged (files only)
	staged, unstaged, totalFiles := 0, 0, 0
	for _, e := range m.allFiles {
		totalFiles++
		if e.staged {
			staged++
		} else {
			unstaged++
		}
	}
	header := fmt.Sprintf("  %d file%s changed (%d staged, %d unstaged)",
		totalFiles, afPlural(totalFiles), staged, unstaged)
	view.WriteString(afHeaderStyle.Render(header) + "\n\n")

	// scrollable window
	maxVisible := m.height - 8
	if maxVisible < 1 {
		maxVisible = 1
	}

	start := m.cursor - maxVisible/2
	if start < 0 {
		start = 0
	}
	end := start + maxVisible
	if end > len(m.entries) {
		end = len(m.entries)
		start = end - maxVisible
		if start < 0 {
			start = 0
		}
	}

	for i := start; i < end; i++ {
		e := m.entries[i]

		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		if e.isDir {
			// render directory row with staging indicator
			dirLabel := e.dir
			if dirLabel == "" || dirLabel == "." {
				dirLabel = "."
			}

			// check if all files in this dir are staged
			allStaged := afAllDirFilesStaged(m.allFiles, e.dir)
			stageTag := afUnstagedStyle.Render("  ")
			if allStaged {
				stageTag = afStagedStyle.Render("  ")
			}

			folderIcon := "󰉋"
			if m.expandedDirs[e.dir] {
				folderIcon = "󰝰" // open folder icon
			}

			if i == m.cursor {
				line := afCursorStyle.Render(fmt.Sprintf("%s%s %s/", cursor, folderIcon, dirLabel))
				view.WriteString(line + stageTag + "\n")
			} else {
				line := fmt.Sprintf("%s%s", cursor, afDirStyle.Render(fmt.Sprintf("%s %s/", folderIcon, dirLabel)))
				view.WriteString(line + stageTag + "\n")
			}
		} else {
			// figure out if this is the last file in its directory group
			isLastInDir := true
			if i+1 < len(m.entries) && !m.entries[i+1].isDir && m.entries[i+1].dir == e.dir {
				isLastInDir = false
			}

			connector := "├─ "
			if isLastInDir {
				connector = "└─ "
			}

			// pick icon and style based on file state
			icon, style := afFileStyle(e)

			// staging indicator
			stageTag := afUnstagedStyle.Render("  ")
			if e.staged {
				stageTag = afStagedStyle.Render("  ")
			}

			// the status code
			statusTag := afHintStyle.Render(fmt.Sprintf(" [%s]", strings.TrimSpace(e.status)))

			styledConnector := afDirStyle.Render(connector)

			if i == m.cursor {
				line := afCursorStyle.Render(fmt.Sprintf("%s%s%s %s", cursor, connector, icon, e.name))
				view.WriteString(line + stageTag + statusTag + "\n")
			} else {
				line := fmt.Sprintf("%s%s%s %s", cursor, styledConnector, icon, style.Render(e.name))
				view.WriteString(line + stageTag + statusTag + "\n")
			}
		}
	}

	shortcuts := []common.Shortcut{
		{Key: "esc/q", Desc: "back"},
		{Key: "space", Desc: "toggle folder"},
		{Key: "a", Desc: "toggle file/folder"},
		{Key: "A", Desc: "stage all"},
	}

	if len(m.entries) > maxVisible {
		pos := fmt.Sprintf("[%d/%d]", m.cursor+1, len(m.entries))
		shortcuts = append(shortcuts, common.Shortcut{Key: "pos", Desc: pos})
	}

	view.WriteString("\n  " + common.RenderShortcuts(shortcuts) + "\n")

	return view.String()
}

// refresh the entry list from git status
func (m *AddFilesModel) refresh() {
	m.allFiles = afGetDirtyFiles()
	m.entries = afBuildEntries(m.allFiles, m.expandedDirs)
	if m.cursor >= len(m.entries) {
		m.cursor = len(m.entries) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// build the list with dir headers + file entries
func afBuildEntries(files []dirtyFile, expandedDirs map[string]bool) []dirtyFile {
	if len(files) == 0 {
		return nil
	}

	var entries []dirtyFile
	lastDir := ""

	for _, f := range files {
		// initialize expanded state if we haven't seen this dir
		if _, ok := expandedDirs[f.dir]; !ok {
			expandedDirs[f.dir] = true // default to expanded
		}

		// insert a dir header when we enter a new directory
		if f.dir != lastDir {
			entries = append(entries, dirtyFile{
				dir:   f.dir,
				name:  f.dir,
				isDir: true,
			})
			lastDir = f.dir
		}

		if expandedDirs[f.dir] {
			entries = append(entries, f)
		}
	}

	return entries
}

// grab all dirty/untracked files from git status --porcelain
func afGetDirtyFiles() []dirtyFile {
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var files []dirtyFile
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}

		status := line[:2]
		rawPath := line[3:]

		// handle renames: "old -> new"
		parts := strings.SplitN(rawPath, " -> ", 2)
		path := parts[len(parts)-1]
		path = strings.Trim(path, "\"")

		f := dirtyFile{
			path:   path,
			name:   filepath.Base(path),
			status: status,
			dir:    filepath.Dir(path),
		}

		// figure out staging state from the status codes
		indexCode := status[0]
		workCode := status[1]

		if status == "??" {
			f.untracked = true
			f.staged = false
		} else if indexCode != ' ' && indexCode != '?' {
			f.staged = true
		}

		if indexCode == 'D' || workCode == 'D' {
			f.deleted = true
		}

		files = append(files, f)
	}

	// sort by dir then name
	sort.Slice(files, func(i, j int) bool {
		if files[i].dir != files[j].dir {
			return files[i].dir < files[j].dir
		}
		return files[i].name < files[j].name
	})

	return files
}

// toggle staging for a single file
func afToggleStaging(f dirtyFile) {
	if f.staged {
		cmd := exec.Command("git", "reset", "HEAD", "--", f.path)
		_ = cmd.Run()
	} else {
		cmd := exec.Command("git", "add", "--", f.path)
		_ = cmd.Run()
	}
}

// toggle staging for all files under a directory
func afToggleDirStaging(allFiles []dirtyFile, targetDir string) {
	// collect all file entries in this dir
	allStaged := true
	var filePaths []string
	for _, f := range allFiles {
		if f.dir == targetDir {
			filePaths = append(filePaths, f.path)
			if !f.staged {
				allStaged = false
			}
		}
	}

	if allStaged {
		// unstage all
		for _, p := range filePaths {
			cmd := exec.Command("git", "reset", "HEAD", "--", p)
			_ = cmd.Run()
		}
	} else {
		// stage all
		for _, p := range filePaths {
			cmd := exec.Command("git", "add", "--", p)
			_ = cmd.Run()
		}
	}
}

// check if all files under the targetDir are staged
func afAllDirFilesStaged(allFiles []dirtyFile, targetDir string) bool {
	hasFiles := false

	for _, f := range allFiles {
		if f.dir == targetDir {
			hasFiles = true
			if !f.staged {
				return false
			}
		}
	}
	return hasFiles
}

// stage all dirty files
func afStageAll() {
	cmd := exec.Command("git", "add", "-A")
	_ = cmd.Run()
}

// pick the right icon and style for a file based on its state
func afFileStyle(f dirtyFile) (string, lipgloss.Style) {
	if f.deleted {
		return "󰮘", afDeletedStyle // nf-md-file_remove (with strike style)
	}
	if f.untracked {
		return "󰝒", afUntrackedStyle // nf-md-file_plus
	}
	if f.staged {
		return "󰄬", afStagedStyle // nf-md-check
	}
	return "󱇧", afUnstagedStyle // nf-md-file_edit
}

// "s" for plural
func afPlural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
