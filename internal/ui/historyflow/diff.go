package historyflow

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rokybeast/zengit/internal/ui/common"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// file ops types from git (must match git's status letters, because git is a-)
const (
	opAdd      = "A"
	opDelete   = "D"
	opModify   = "M"
	opRename   = "R"
	opCopy     = "C"
	opType     = "T" // type change (mode); or, chmod for linux chuds
	opUnmerged = "U"
)

// sexy looking nf iconssss
const (
	nfAdd      = "󰝒"
	nfDelete   = "󰮘"
	nfModify   = "󱇧"
	nfRename   = "󰏫"
	nfCopy     = "󰆏"
	nfMode     = "󰌾"
	nfConflict = "󰘭"
	nfBinary   = "󰈔"
	nfSymlink  = "󰌹"
)

var (
	diffTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(common.ColorFrostBlue).
			PaddingLeft(1)

	diffFileStyle = lipgloss.NewStyle().
			Foreground(common.ColorFrostBlue).
			Bold(true)

	diffDividerStyle = lipgloss.NewStyle().
				Foreground(common.ColorBgDiffLine)

	diffBodyStyle = lipgloss.NewStyle().
			Foreground(common.ColorSnowDark)

	diffAddLineStyle = lipgloss.NewStyle().
				Foreground(common.ColorGreen).
				Background(common.ColorBgDiffAdd)

	diffDelLineStyle = lipgloss.NewStyle().
				Foreground(common.ColorRed).
				Background(common.ColorBgDiffDelete)

	diffHintStyle = lipgloss.NewStyle().
			Foreground(common.ColorMutedGray).
			PaddingLeft(1)

	diffMutedStyle = lipgloss.NewStyle().
			Foreground(common.ColorMutedGray)

	diffTreeAddStyle = lipgloss.NewStyle().
				Foreground(common.ColorGreen)

	diffTreeDelStyle = lipgloss.NewStyle().
				Foreground(common.ColorRed)

	diffTreeGlyphStyle = lipgloss.NewStyle().
				Foreground(common.ColorMutedGray)
)

// a single file entry in a commit
type fileEntry struct {
	op       string // A, D, M, R, C, T, U
	path     string
	oldPath  string // for renames/copies
	addLines int
	delLines int
	diffText string // parsed diff lines for this file
	isBinary bool
}

// the diff sub-model
type DiffModel struct {
	hash     string
	message  string
	files    []fileEntry
	viewport viewport.Model
	width    int
	height   int
	detailed bool // true = detailed modede, false = summarize mode
	ready    bool
}

// back from diff to graph
type diffBackMsg struct{}

func newDiffModel(hash, message string, width, height int) DiffModel {
	files := parseCommitFiles(hash)
	vp := viewport.New(width, height-3)
	vp.YPosition = 3

	m := DiffModel{
		hash:     hash,
		message:  message,
		files:    files,
		viewport: vp,
		width:    width,
		height:   height,
		detailed: true,
		ready:    true,
	}
	m.rebuildContent()
	return m
}

func (m DiffModel) Init() tea.Cmd {
	return nil
}

func (m DiffModel) Update(msg tea.Msg) (DiffModel, tea.Cmd) {
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
			return m, func() tea.Msg { return diffBackMsg{} }
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

func (m DiffModel) View() string {
	shortHash := m.hash
	if len(shortHash) > 7 {
		shortHash = shortHash[:7]
	}

	title := diffTitleStyle.Render(
		fmt.Sprintf("commit: '%s' [%s]", m.message, shortHash),
	)

	shortcuts := []common.Shortcut{
		{Key: "esc/q", Desc: "back"},
		{Key: "j/k", Desc: "scroll"},
	}

	if m.detailed {
		shortcuts = append(shortcuts, common.Shortcut{Key: "o", Desc: "summarized view"})
	} else {
		shortcuts = append(shortcuts, common.Shortcut{Key: "o", Desc: "detailed view"})
	}

	footer := "\n  " + common.RenderShortcuts(shortcuts) + "\n"

	return title + "\n\n" + m.viewport.View() + footer
}

func (m *DiffModel) rebuildContent() {
	var content string
	if m.detailed {
		content = m.buildDetailed()
	} else {
		content = m.buildSummarized()
	}
	m.viewport.SetContent(content)
	m.viewport.GotoTop()
}

func (m *DiffModel) divider() string {
	w := m.width
	if w < 1 {
		w = 80
	}
	return diffDividerStyle.Render(strings.Repeat("─", w))
}

// detailed view: full diffs with colored +/- lines
func (m *DiffModel) buildDetailed() string {
	var b strings.Builder
	w := m.width
	if w < 1 {
		w = 80
	}

	for _, f := range m.files {
		b.WriteString(m.divider() + "\n")

		switch f.op {
		case opAdd:
			b.WriteString(diffFileStyle.Render(nfAdd+" new file: "+f.path) + "\n")
			if f.diffText == "" {
				b.WriteString("\n" + diffBodyStyle.Render("empty") + "\n")
			} else {
				b.WriteString("\n")
				b.WriteString(m.renderDiffLines(f.diffText, w))
			}
		case opDelete:
			b.WriteString(diffFileStyle.Render(nfDelete+" delete file: "+f.path) + "\n")
			if f.delLines > 0 {
				b.WriteString(diffDelLineStyle.Render(
					fmt.Sprintf("- removed %d %s", f.delLines, pluralLine(f.delLines))) + "\n")
			}
		case opModify:
			if f.isBinary {
				b.WriteString(diffFileStyle.Render(nfBinary+" binfile: "+f.path) + "\n")
				b.WriteString("\n" + diffBodyStyle.Render("zengit cannot load raw binary files") + "\n")
			} else {
				b.WriteString(diffFileStyle.Render(nfModify+" edit file: "+f.path) + "\n")
				if f.diffText != "" {
					b.WriteString("\n")
					b.WriteString(m.renderDiffLines(f.diffText, w))
				}
			}
		case opRename:
			label := f.path
			if f.oldPath != "" {
				label = f.oldPath + " → " + f.path
			}
			b.WriteString(diffFileStyle.Render(nfRename+" rename file: "+label) + "\n")
			if f.diffText != "" {
				b.WriteString("\n")
				b.WriteString(m.renderDiffLines(f.diffText, w))
			}
		case opCopy:
			b.WriteString(diffFileStyle.Render(nfCopy+" copy: "+f.path) + "\n")
			b.WriteString("\n" + diffBodyStyle.Render("copied file: "+f.path) + "\n")
		case opType:
			b.WriteString(diffFileStyle.Render(nfMode+" mode: "+f.path) + "\n")
			b.WriteString("\n" + diffBodyStyle.Render("permission change") + "\n")
		case opUnmerged:
			b.WriteString(diffFileStyle.Render(nfConflict+" merge conflict: "+f.path) + "\n")
			b.WriteString("\n" + diffBodyStyle.Render("detected merge conflict in: "+f.path) + "\n")
		default:
			b.WriteString(diffFileStyle.Render(" "+f.op+": "+f.path) + "\n")
		}
	}
	b.WriteString(m.divider() + "\n")

	return b.String()
}

// render each diff lines
func (m *DiffModel) renderDiffLines(raw string, w int) string {
	var b strings.Builder
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		if strings.Contains(line, " + | ") {
			b.WriteString(diffAddLineStyle.Render(line) + "\n")
		} else if strings.Contains(line, " - | ") {
			b.WriteString(diffDelLineStyle.Render(line) + "\n")
		} else {
			b.WriteString(diffBodyStyle.Render(line) + "\n")
		}
	}
	return b.String()
}

// summarized view
func (m *DiffModel) buildSummarized() string {
	var b strings.Builder

	for _, f := range m.files {
		switch f.op {
		case opAdd:
			b.WriteString(diffFileStyle.Render(nfAdd+" new file: "+f.path) + "\n")
			if f.addLines == 0 && f.delLines == 0 {
				b.WriteString(diffTreeGlyphStyle.Render("└ ") + diffMutedStyle.Render("empty file") + "\n")
			} else if f.addLines > 0 {
				b.WriteString(diffTreeGlyphStyle.Render("└ ") +
					diffTreeAddStyle.Render(fmt.Sprintf("+%d %s", f.addLines, pluralLine(f.addLines))) + "\n")
			}
		case opDelete:
			b.WriteString(diffFileStyle.Render(nfDelete+" delete file: "+f.path) + "\n")
			if f.delLines > 0 {
				b.WriteString(diffTreeGlyphStyle.Render("└ ") +
					diffTreeDelStyle.Render(fmt.Sprintf("-%d %s", f.delLines, pluralLine(f.delLines))) + "\n")
			}
		case opModify:
			if f.isBinary {
				b.WriteString(diffFileStyle.Render(nfBinary+" binfile: "+f.path) + "\n")
			} else {
				b.WriteString(diffFileStyle.Render(nfModify+" edit file: "+f.path) + "\n")
				if f.addLines > 0 && f.delLines > 0 {
					b.WriteString(diffTreeGlyphStyle.Render("├ ") +
						diffTreeAddStyle.Render(fmt.Sprintf("+%d %s", f.addLines, pluralLine(f.addLines))) + "\n")
					b.WriteString(diffTreeGlyphStyle.Render("└ ") +
						diffTreeDelStyle.Render(fmt.Sprintf("-%d %s", f.delLines, pluralLine(f.delLines))) + "\n")
				} else if f.addLines > 0 {
					b.WriteString(diffTreeGlyphStyle.Render("└ ") +
						diffTreeAddStyle.Render(fmt.Sprintf("+%d %s", f.addLines, pluralLine(f.addLines))) + "\n")
				} else if f.delLines > 0 {
					b.WriteString(diffTreeGlyphStyle.Render("└ ") +
						diffTreeDelStyle.Render(fmt.Sprintf("-%d %s", f.delLines, pluralLine(f.delLines))) + "\n")
				}
			}
		case opRename:
			label := f.path
			if f.oldPath != "" {
				label = f.oldPath + " → " + f.path
			}
			b.WriteString(diffFileStyle.Render(nfRename+" rename file: "+label) + "\n")
		case opCopy:
			b.WriteString(diffFileStyle.Render(nfCopy+" copy: "+f.path) + "\n")
		case opType:
			b.WriteString(diffFileStyle.Render(nfMode+" mode: "+f.path) + "\n")
		case opUnmerged:
			b.WriteString(diffFileStyle.Render(nfConflict+" merge conflict: "+f.path) + "\n")
		default:
			b.WriteString(diffFileStyle.Render(" "+f.op+": "+f.path) + "\n")
		}
	}
	return b.String()
}

// litle helper for number detection (its small)
func pluralLine(n int) string {
	if n == 1 {
		return "line"
	}
	return "lines"
}

// parse the files changed in a commit using git show and git diff
func parseCommitFiles(hash string) []fileEntry {
	// get name-status for operation types
	nsCmd := exec.Command("git", "show", "--format=", "--name-status", "-M", "-C", hash)
	nsOut, err := nsCmd.Output()
	if err != nil {
		return nil
	}

	// get numstat for line counts
	numCmd := exec.Command("git", "show", "--format=", "--numstat", hash)
	numOut, _ := numCmd.Output()
	numStats := parseNumstat(string(numOut))

	// get the full diff
	diffCmd := exec.Command("git", "diff-tree", "-p", "--no-commit-id", "-M", "-C", hash)
	diffOut, _ := diffCmd.Output()
	diffMap := parseDiffOutput(string(diffOut))

	var files []fileEntry
	lines := strings.Split(strings.TrimSpace(string(nsOut)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 2 {
			continue
		}

		op := parts[0]
		path := parts[1]
		oldPath := ""

		// rename/copy have a similar level like R100 or C095 [yes, i used ai for this logic]
		if len(op) > 1 && (op[0] == 'R' || op[0] == 'C') {
			if len(parts) >= 3 {
				oldPath = parts[1]
				path = parts[2]
			}
			op = string(op[0])
		}

		f := fileEntry{
			op:      op,
			path:    path,
			oldPath: oldPath,
		}

		// pull numstat counts
		if ns, ok := numStats[path]; ok {
			f.addLines = ns.add
			f.delLines = ns.del
			f.isBinary = ns.binary
		}

		// pull diff text
		lookupKey := path
		if oldPath != "" {
			lookupKey = path
		}
		if dt, ok := diffMap[lookupKey]; ok {
			f.diffText = dt
		}
		// fallback: try old path for renames
		if f.diffText == "" && oldPath != "" {
			if dt, ok := diffMap[oldPath]; ok {
				f.diffText = dt
			}
		}

		files = append(files, f)
	}
	return files
}

type numstatEntry struct {
	add    int
	del    int
	binary bool
}

func parseNumstat(raw string) map[string]numstatEntry {
	result := make(map[string]numstatEntry)
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 {
			continue
		}
		path := parts[2]
		// handle renames: "old -> new" or "{dir/old -> dir/new}"
		if idx := strings.Index(path, " => "); idx >= 0 {
			// take the new name (right side)
			path = strings.TrimSpace(path[idx+4:])
			path = strings.TrimRight(path, "}")
		}

		if parts[0] == "-" && parts[1] == "-" {
			result[path] = numstatEntry{binary: true}
		} else {
			a, _ := strconv.Atoi(parts[0])
			d, _ := strconv.Atoi(parts[1])
			result[path] = numstatEntry{add: a, del: d}
		}
	}
	return result
}

// parse full diff output into eachfile chunks keyed by filename
func parseDiffOutput(raw string) map[string]string {
	result := make(map[string]string)
	// split on "diff --git" boundaries
	chunks := strings.Split(raw, "diff --git ")
	for _, chunk := range chunks {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		// extract the filename from the +++ b/path line
		var fname string
		lines := strings.Split(chunk, "\n")
		for _, l := range lines {
			if strings.HasPrefix(l, "+++ b/") {
				fname = strings.TrimPrefix(l, "+++ b/")
				break
			}
			// deleted files show +++ /dev/null, try --- a/ instead
			if strings.HasPrefix(l, "--- a/") && fname == "" {
				fname = strings.TrimPrefix(l, "--- a/")
			}
		}
		if fname == "" || fname == "/dev/null" {
			// extract from first line: a/path b/path
			firstLine := lines[0]
			parts := strings.SplitN(firstLine, " ", 2)
			if len(parts) == 2 {
				fname = strings.TrimPrefix(parts[1], "b/")
			}
		}

		// collect only the +/- lines (skip headers and @@ hunks)
		var diffLines []string
		lineNum := 0
		for _, l := range lines {
			if strings.HasPrefix(l, "@@") {
				// parse the line number from "@@ -a,b +c,d @@"
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
				diffLines = append(diffLines, fmt.Sprintf("%d + | %s", lineNum, l[1:]))
				lineNum++
			} else if strings.HasPrefix(l, "-") {
				diffLines = append(diffLines, fmt.Sprintf("%d - | %s", lineNum, l[1:]))
				// don't increment lineNum for deletions (they were in old file)
			} else if strings.HasPrefix(l, " ") {
				lineNum++
			}
		}
		if fname != "" {
			result[fname] = strings.Join(diffLines, "\n")
		}
	}
	return result
}

// parse the starting line number from a hunk header like "@@ -1,3 +1,5 @@z"
func parseHunkStart(hunk string) int {
	// find +N in the @@ ... @@ line
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
