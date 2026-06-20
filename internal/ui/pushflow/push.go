package pushflow

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"gitty/internal/git"
	"gitty/internal/ui/common"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type pushStep int

const (
	stepCommits pushStep = iota
	stepRemotes
	stepPushing
	stepDone
)

type BackMsg struct{}

type commitItem struct {
	hash     string
	short    string
	message  string
	selected bool
	implicit bool
	explicit bool
}

type remoteItem struct {
	name     string
	url      string
	upToDate bool
	selected bool
}

type pushResultMsg struct {
	err    error
	stderr string
}

type tickMsg struct{}

type Model struct {
	step    pushStep
	commits []commitItem
	remotes []remoteItem
	cursor  int
	spinner spinner.Model
	pushing bool
	success bool
	pushErr string
	width   int
	height  int
}

var (
	explicitCheckStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#A3BE8C")).Bold(true)
	implicitCheckStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#A3BE8C")).Faint(true)
	uncheckedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#4C566A"))
	hashStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#81A1C1"))
	msgStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("#eceff4"))
	cursorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")).Bold(true)
	titleStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#88c0d0")).PaddingLeft(4)

	pushSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A3BE8C")). // nord green
				Bold(true).
				PaddingLeft(4)

	pushDetailStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d8dee9")). // nord snow
			PaddingLeft(4)

	pushHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4c566a")). // nord muted gray
			PaddingLeft(4).
			MarginTop(1)

	pushErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BF616A")). // nord red
			Bold(true).
			PaddingLeft(4)
)

func New(width, height int) Model {
	commits := getUnpushedCommits()
	remotes := getRemotes()

	var step pushStep
	if len(commits) == 0 {
		step = stepDone
	} else {
		step = stepCommits
		if len(commits) == 1 {
			commits[0].selected = true
			commits[0].explicit = true
		}
	}

	s := spinner.New()
	s.Spinner = spinner.Spinner{
		Frames: []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"},
		FPS:    time.Millisecond * 100,
	}
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#88c0d0"))

	return Model{
		step:    step,
		commits: commits,
		remotes: remotes,
		spinner: s,
		width:   width,
		height:  height,
	}
}

// bgproc init
func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, m.spinner.Tick)
	if m.step == stepPushing {
		cmds = append(cmds, m.startPush())
	}
	return tea.Batch(cmds...)
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
		case "esc":
			return m, func() tea.Msg { return BackMsg{} }
		}

		switch m.step {
		case stepCommits:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.commits)-1 {
					m.cursor++
				}
			case " ":
				if len(m.commits) > 0 {
					m.toggleCommit(m.cursor)
				}
			case "enter":
				anySelected := false
				for _, c := range m.commits {
					if c.selected {
						anySelected = true
						break
					}
				}
				if !anySelected {
					return m, nil
				}
				if len(m.remotes) > 1 {
					m.step = stepRemotes
					m.cursor = 0
				} else {
					m.step = stepPushing
					return m, m.startPush()
				}
			}

		case stepRemotes:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.remotes)-1 {
					m.cursor++
				}
			case " ":
				if len(m.remotes) > 0 {
					m.selectRemote(m.cursor)
				}
			case "enter":
				if len(m.remotes) > 0 {
					m.selectRemote(m.cursor)
				}
				m.step = stepPushing
				return m, m.startPush()
			}

		case stepDone:
			switch msg.String() {
			case "enter", "q":
				return m, func() tea.Msg { return BackMsg{} }
			}
		}

	case pushResultMsg:
		m.pushing = false
		if msg.err != nil {
			m.pushErr = msg.stderr
			if m.pushErr == "" {
				m.pushErr = msg.err.Error()
			}
			m.step = stepDone
			return m, nil
		}
		m.success = true
		m.step = stepDone
		return m, nil
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	var view strings.Builder
	indent := "  "

	switch m.step {
	case stepCommits:
		if len(m.commits) > 1 {
			view.WriteString("\n" + titleStyle.Render("select commits to push") + "\n\n")
		} else {
			view.WriteString("\n")
		}
		for i, c := range m.commits {
			cursorPrefix := "   "
			if i == m.cursor {
				cursorPrefix = cursorStyle.Render("-> ")
			}
			checkStr := "[ ]"
			if c.selected {
				if c.implicit {
					checkStr = implicitCheckStyle.Render("[x]")
				} else {
					checkStr = explicitCheckStyle.Render("[x]")
				}
			} else {
				checkStr = uncheckedStyle.Render("[ ]")
			}
			view.WriteString(fmt.Sprintf("%s%s%s %s - '%s'\n",
				indent,
				cursorPrefix,
				checkStr,
				hashStyle.Render("("+c.short+")"),
				msgStyle.Render(c.message),
			))
		}
		shortcuts := []common.Shortcut{
			{Key: "esc", Desc: "back"},
			{Key: "space", Desc: "toggle"},
			{Key: "enter", Desc: "push / next"},
		}
		view.WriteString("\n  " + common.RenderShortcuts(shortcuts) + "\n")

	case stepRemotes:
		if len(m.remotes) > 1 {
			view.WriteString("\n" + titleStyle.Render("push where?") + "\n\n")
		} else {
			view.WriteString("\n")
		}
		for i, r := range m.remotes {
			cursorPrefix := "   "
			if i == m.cursor {
				cursorPrefix = cursorStyle.Render("-> ")
			}
			checkStr := "[ ]"
			if r.selected {
				checkStr = explicitCheckStyle.Render("[x]")
			} else {
				checkStr = uncheckedStyle.Render("[ ]")
			}
			statusText := ""
			if r.upToDate {
				statusText = lipgloss.NewStyle().Foreground(lipgloss.Color("#A3BE8C")).Render("[up to date]")
			} else {
				statusText = lipgloss.NewStyle().Foreground(lipgloss.Color("#BF616A")).Render("[not up to date]")
			}
			urlStyle := lipgloss.NewStyle().Faint(true)
			remoteNameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#eceff4"))
			view.WriteString(fmt.Sprintf("%s%s%s %s %s %s\n",
				indent,
				cursorPrefix,
				checkStr,
				remoteNameStyle.Render(r.name),
				urlStyle.Render("("+r.url+")"),
				statusText,
			))
		}
		shortcuts := []common.Shortcut{
			{Key: "esc", Desc: "back"},
			{Key: "space/enter", Desc: "select"},
		}
		view.WriteString("\n  " + common.RenderShortcuts(shortcuts) + "\n")

	case stepPushing:
		view.WriteString("\n")
		numCommits := 0
		var selectedHashes []string
		for _, c := range m.commits {
			if c.selected {
				numCommits++
				selectedHashes = append(selectedHashes, c.short)
			}
		}
		hashList := strings.Join(selectedHashes, ", ")
		boldN := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%d", numCommits))

		selectedRemote := ""
		for _, r := range m.remotes {
			if r.selected {
				selectedRemote = r.name
				break
			}
		}
		if selectedRemote == "" && len(m.remotes) > 0 {
			selectedRemote = m.remotes[0].name
		}
		boldRemote := lipgloss.NewStyle().Bold(true).Render(selectedRemote)
		arrow := lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")).Bold(true).Render("->")

		if m.pushing {
			spinnerStr := m.spinner.View()
			view.WriteString(fmt.Sprintf("%s%s Pushing %s Commits (%s) %s %s...\n",
				indent,
				spinnerStr,
				boldN,
				hashList,
				arrow,
				boldRemote,
			))
		}
	case stepDone:
		view.WriteString(m.viewDone())
	}

	return view.String()
}

// show the commit options
func (m *Model) toggleCommit(idx int) {
	if idx == 0 {
		targetState := !m.commits[0].selected
		for k := range m.commits {
			m.commits[k].explicit = targetState
			m.commits[k].selected = targetState
			m.commits[k].implicit = false
		}
	} else {
		m.commits[idx].explicit = !m.commits[idx].explicit
		m.recalculateSelection()
	}
}

// recalculate contiguous selection boundaries
func (m *Model) recalculateSelection() {
	minIdx := -1
	maxIdx := -1
	for i, c := range m.commits {
		if c.explicit {
			if minIdx == -1 {
				minIdx = i
			}
			maxIdx = i
		}
	}

	if minIdx == -1 {
		for i := range m.commits {
			m.commits[i].selected = false
			m.commits[i].implicit = false
		}
		return
	}

	for i := range m.commits {
		if i >= minIdx && i <= maxIdx {
			m.commits[i].selected = true
			if !m.commits[i].explicit {
				m.commits[i].implicit = true
			} else {
				m.commits[i].implicit = false
			}
		} else {
			m.commits[i].selected = false
			m.commits[i].implicit = false
			m.commits[i].explicit = false
		}
	}
}

// select target remote and clear others
func (m *Model) selectRemote(idx int) {
	for i := range m.remotes {
		m.remotes[i].selected = (i == idx)
	}
}

// start the git push process in bg
func (m *Model) startPush() tea.Cmd {
	m.pushing = true
	m.success = false
	m.pushErr = ""

	selectedRemote := ""
	for _, r := range m.remotes {
		if r.selected {
			selectedRemote = r.name
			break
		}
	}
	if selectedRemote == "" && len(m.remotes) > 0 {
		selectedRemote = m.remotes[0].name
	}

	allSelected := true
	for _, c := range m.commits {
		if !c.selected {
			allSelected = false
			break
		}
	}

	var oldestHash string
	for i := len(m.commits) - 1; i >= 0; i-- {
		if m.commits[i].selected {
			oldestHash = m.commits[i].hash
			break
		}
	}

	branch := git.CurrentBranch()
	return runPushCmd(selectedRemote, branch, oldestHash, allSelected)
}

// get list of unpushed commits
func getUnpushedCommits() []commitItem {
	var cmd *exec.Cmd
	upstreamCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "@{u}")
	if err := upstreamCmd.Run(); err == nil {
		cmd = exec.Command("git", "log", "@{u}..HEAD", "--format=%H %s")
	} else {
		cmd = exec.Command("git", "log", "HEAD", "--not", "--remotes", "--format=%H %s")
	}
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var commits []commitItem
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		hash := parts[0]
		msg := ""
		if len(parts) > 1 {
			msg = parts[1]
		}
		short := hash
		if len(short) > 7 {
			short = short[:7]
		}
		commits = append(commits, commitItem{
			hash:    hash,
			short:   short,
			message: msg,
		})
	}
	return commits
}

// get configured remotes with statuses
func getRemotes() []remoteItem {
	cmd := exec.Command("git", "remote", "-v")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	remoteMap := make(map[string]string)
	var names []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			name := fields[0]
			url := fields[1]
			if _, exists := remoteMap[name]; !exists {
				remoteMap[name] = url
				names = append(names, name)
			}
		}
	}

	branch := git.CurrentBranch()
	var remotes []remoteItem
	for _, name := range names {
		url := remoteMap[name]
		upToDate := true
		countCmd := exec.Command("git", "rev-list", "--count", fmt.Sprintf("%s/%s..HEAD", name, branch))
		countOut, err := countCmd.Output()
		if err == nil {
			var count int
			_, _ = fmt.Sscanf(strings.TrimSpace(string(countOut)), "%d", &count)
			if count > 0 {
				upToDate = false
			}
		}
		remotes = append(remotes, remoteItem{
			name:     name,
			url:      url,
			upToDate: upToDate,
		})
	}

	if len(remotes) > 0 {
		foundOrigin := false
		for i, r := range remotes {
			if r.name == "origin" {
				remotes[i].selected = true
				foundOrigin = true
				break
			}
		}
		if !foundOrigin {
			remotes[0].selected = true
		}
	}

	return remotes
}

// run push command syncly (not a word, im just lazy-) in task cmd
func runPushCmd(remote string, branch string, oldestHash string, allSelected bool) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		if allSelected || oldestHash == "" {
			cmd = exec.Command("git", "push", remote, "HEAD")
		} else {
			refspec := fmt.Sprintf("%s:refs/heads/%s", oldestHash, branch)
			cmd = exec.Command("git", "push", remote, refspec)
		}
		var stderr strings.Builder
		cmd.Stderr = &stderr
		err := cmd.Run()
		return pushResultMsg{
			err:    err,
			stderr: stderr.String(),
		}
	}
}

func (m Model) viewDone() string {
	if m.pushErr != "" {
		errMsg := pushErrorStyle.Render("push failed!")
		detail := pushDetailStyle.Render(m.pushErr)
		
		shortcuts := []common.Shortcut{
			{Key: "esc", Desc: "back"},
		}
		footer := "\n  " + common.RenderShortcuts(shortcuts)
		
		return lipgloss.JoinVertical(lipgloss.Left, "", errMsg, "", detail, "", footer)
	}

	repoName := git.RepoName()
	branch := git.CurrentBranch()

	hasChanges, hasPushes := git.CheckRepoStatus()
	if !hasChanges && !hasPushes {
		messageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#81a1c1")).Bold(true)
		msg := messageStyle.Render(fmt.Sprintf("󰳏 %s/%s is clean and nothing is left, are you done for the day? hope not ;)", repoName, branch))
		
		shortcuts := []common.Shortcut{
			{Key: "esc/enter/q", Desc: "back"},
		}
		footer := common.RenderShortcuts(shortcuts)
		
		fullText := lipgloss.JoinVertical(lipgloss.Center, msg, "", footer)
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fullText)
	}

	var header string
	var detail string

	if len(m.commits) == 0 {
		header = pushSuccessStyle.Render(
			fmt.Sprintf("everything up-to-date on [\uf126 %s/%s]", repoName, branch),
		)
		detail = pushDetailStyle.Render("no unpushed commits found.")
	} else {
		numCommits := 0
		var selectedHashes []string
		for _, c := range m.commits {
			if c.selected {
				numCommits++
				selectedHashes = append(selectedHashes, c.short)
			}
		}
		hashList := strings.Join(selectedHashes, ", ")

		selectedRemote := ""
		for _, r := range m.remotes {
			if r.selected {
				selectedRemote = r.name
				break
			}
		}
		if selectedRemote == "" && len(m.remotes) > 0 {
			selectedRemote = m.remotes[0].name
		}

		header = pushSuccessStyle.Render(
			fmt.Sprintf("pushed to [\uf126 %s/%s]", repoName, branch),
		)
		detail = pushDetailStyle.Render(
			fmt.Sprintf("pushed %d commit(s) (%s) -> %s", numCommits, hashList, selectedRemote),
		)
	}

	shortcuts := []common.Shortcut{
		{Key: "esc", Desc: "menu"},
	}
	footer := "\n  " + common.RenderShortcuts(shortcuts)

	return lipgloss.JoinVertical(lipgloss.Left, "", header, "", detail, "", footer)
}
