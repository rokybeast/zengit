package ui

import (
	"fmt"
	"os"

	"github.com/rokybeast/zengit/internal/git"
	"github.com/rokybeast/zengit/internal/ui/about"
	"github.com/rokybeast/zengit/internal/ui/common"
	"github.com/rokybeast/zengit/internal/ui/commitflow"
	"github.com/rokybeast/zengit/internal/ui/historyflow"
	"github.com/rokybeast/zengit/internal/ui/initflow"
	"github.com/rokybeast/zengit/internal/ui/menu"
	"github.com/rokybeast/zengit/internal/ui/pushflow"
	"github.com/rokybeast/zengit/internal/ui/treeflow"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateNoGit state = iota
	stateGit
	stateInitRepo
	stateNav
	stateAbout
	stateCommit
	stateTree
	stateMessage
	statePush
	stateHistory
	stateAddFiles
)

type Model struct {
	state        state
	prevState    state // where to go back to from about/sub-screens
	noGit        menu.NoGitModel
	gitMenu      menu.GitModel
	initFlow     initflow.Model
	commitFlow   commitflow.Model
	treeFlow     treeflow.Model
	pushFlow     pushflow.Model
	navFlow      treeflow.NoGitModel
	historyFlow  historyflow.Model
	addFilesFlow treeflow.AddFilesModel
	about        about.Model
	quitting     bool
	width        int
	height       int
	message      string
}

// make a new fresh model and detect the git repo to pick the first state
func New() Model {
	cwd, _ := os.Getwd()
	s := stateNoGit
	if git.IsRepo(cwd) {
		s = stateGit
	}

	m := Model{
		state: s,
		noGit: menu.NewNoGit(),
	}

	// boot the git menu right away if we're in a repo
	if s == stateGit {
		m.gitMenu = menu.NewGit(0, 0)
	}

	return m
}

func (m Model) Init() tea.Cmd {
	switch m.state {
	case stateNoGit:
		return m.noGit.Init()
	case stateGit:
		return m.gitMenu.Init()
	case stateInitRepo:
		return m.initFlow.Init()
	case stateNav:
		return m.navFlow.Init()
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case menu.ChoiceMsg:
		return m.handleNoGitOption(msg)
	case menu.GitChoiceMsg:
		return m.handleGitOption(msg)
	case initflow.DoneMsg:
		_ = git.InitRepo(msg.GitIgnore, msg.License)
		m.state = stateGit
		m.gitMenu = menu.NewGit(m.width, m.height)
		return m, nil
	case treeflow.PickedMsg:
		// change into the selected repo directory
		_ = os.Chdir(msg.Path)
		m.state = stateGit
		m.gitMenu = menu.NewGit(m.width, m.height)
		return m, nil
	case about.BackMsg, commitflow.BackMsg, treeflow.BackMsg, pushflow.BackMsg, historyflow.BackMsg:
		m.state = m.prevState
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		if m.state == stateMessage {
			switch msg.String() {
			case "esc", "enter", "q":
				m.state = m.prevState
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	var cmd tea.Cmd
	switch m.state {
	case stateNoGit:
		var updated tea.Model
		updated, cmd = m.noGit.Update(msg)
		m.noGit = updated.(menu.NoGitModel)
	case stateGit:
		var updated tea.Model
		updated, cmd = m.gitMenu.Update(msg)
		m.gitMenu = updated.(menu.GitModel)
	case stateInitRepo:
		var updated tea.Model
		updated, cmd = m.initFlow.Update(msg)
		m.initFlow = updated.(initflow.Model)
	case stateNav:
		var updated tea.Model
		updated, cmd = m.navFlow.Update(msg)
		m.navFlow = updated.(treeflow.NoGitModel)
	case stateAbout:
		var updated tea.Model
		updated, cmd = m.about.Update(msg)
		m.about = updated.(about.Model)
	case stateCommit:
		var updated tea.Model
		updated, cmd = m.commitFlow.Update(msg)
		m.commitFlow = updated.(commitflow.Model)
	case stateTree:
		var updated tea.Model
		updated, cmd = m.treeFlow.Update(msg)
		m.treeFlow = updated.(treeflow.Model)
	case statePush:
		var updated tea.Model
		updated, cmd = m.pushFlow.Update(msg)
		m.pushFlow = updated.(pushflow.Model)
	case stateHistory:
		var updated tea.Model
		updated, cmd = m.historyFlow.Update(msg)
		m.historyFlow = updated.(historyflow.Model)
	case stateAddFiles:
		var updated tea.Model
		updated, cmd = m.addFilesFlow.Update(msg)
		m.addFilesFlow = updated.(treeflow.AddFilesModel)
	}

	return m, cmd
}

// renders the active sub-model
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	switch m.state {
	case stateNoGit:
		return m.noGit.View()
	case stateInitRepo:
		return m.initFlow.View()
	case stateNav:
		return m.navFlow.View()
	case stateAbout:
		return m.about.View()
	case stateGit:
		return m.gitMenu.View()
	case stateCommit:
		return m.commitFlow.View()
	case stateTree:
		return m.treeFlow.View()
	case statePush:
		return m.pushFlow.View()
	case stateHistory:
		return m.historyFlow.View()
	case stateAddFiles:
		return m.addFilesFlow.View()
	case stateMessage:
		return m.viewMessage()
	}

	return ""
}

// handle selections from the no-git menu by id
func (m Model) handleNoGitOption(msg menu.ChoiceMsg) (tea.Model, tea.Cmd) {
	switch msg.ID {
	case menu.IDInitRepo:
		m.state = stateInitRepo
		m.initFlow = initflow.New(m.width, m.height)
		return m, m.initFlow.Init()
	case menu.IDNavigate:
		m.state = stateNav
		m.navFlow = treeflow.NewNoGit(m.width, m.height)
		return m, m.navFlow.Init()
	case menu.IDAbout:
		m.prevState = stateNoGit
		m.state = stateAbout
		m.about = about.New()
		return m, m.about.Init()
	case menu.IDQuit:
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

// handle selections from the git repo menu by id
func (m Model) handleGitOption(msg menu.GitChoiceMsg) (tea.Model, tea.Cmd) {
	switch msg.ID {
	case menu.IDAbout:
		m.prevState = stateGit
		m.state = stateAbout
		m.about = about.New()
		return m, m.about.Init()
	case menu.IDCommit:
		hasChanges, hasPushes := git.CheckRepoStatus()
		if !hasChanges {
			m.prevState = stateGit
			m.state = stateMessage
			repo := git.RepoName()
			branch := git.CurrentBranch()
			if hasPushes {
				m.message = fmt.Sprintf("󰳏 %s/%s is clean, now, just push 'em! great job completing it! i think..", repo, branch)
			} else {
				m.message = fmt.Sprintf("󰳏 %s/%s is clean and nothing is left, are you done for the day? hope not ;)", repo, branch)
			}
			return m, nil
		}
		// there are changes but nothing staged yet
		if !git.HasStagedFiles() {
			m.prevState = stateGit
			m.state = stateMessage
			repo := git.RepoName()
			branch := git.CurrentBranch()
			m.message = fmt.Sprintf("󰳏 %s/%s has no files added yet, add them, and come here >:(", repo, branch)
			return m, nil
		}
		m.prevState = stateGit
		m.state = stateCommit
		m.commitFlow = commitflow.New(m.width, m.height)
		return m, m.commitFlow.Init()
	case menu.IDAddFiles:
		hasChanges, hasPushes := git.CheckRepoStatus()
		if !hasChanges {
			m.prevState = stateGit
			m.state = stateMessage
			repo := git.RepoName()
			branch := git.CurrentBranch()
			if hasPushes {
				m.message = fmt.Sprintf("󰳏 %s/%s is clean, now, just push 'em! great job completing it! i think..", repo, branch)
			} else {
				m.message = fmt.Sprintf("󰳏 %s/%s is clean and nothing is left, are you done for the day? hope not ;)", repo, branch)
			}
			return m, nil
		}
		m.prevState = stateGit
		m.state = stateAddFiles
		m.addFilesFlow = treeflow.NewAddFiles(m.width, m.height)
		return m, m.addFilesFlow.Init()
	case menu.IDTree:
		m.prevState = stateGit
		m.state = stateTree
		m.treeFlow = treeflow.New(m.width, m.height, false)
		return m, m.treeFlow.Init()
	case menu.IDPush:
		hasChanges, hasPushes := git.CheckRepoStatus()
		if !hasPushes {
			m.prevState = stateGit
			m.state = stateMessage
			repo := git.RepoName()
			branch := git.CurrentBranch()
			if hasChanges {
				m.message = fmt.Sprintf("󰳏 %s/%s has uncommitted changes, but no pushes are left! get back to work!", repo, branch)
			} else {
				m.message = fmt.Sprintf("󰳏 %s/%s is clean and nothing is left, are you done for the day? hope not ;)", repo, branch)
			}
			return m, nil
		}
		m.prevState = stateGit
		m.state = statePush
		m.pushFlow = pushflow.New(m.width, m.height)
		return m, m.pushFlow.Init()
	case menu.IDHistory:
		m.prevState = stateGit
		m.state = stateHistory
		m.historyFlow = historyflow.New(m.width, m.height)
		return m, m.historyFlow.Init()
	case menu.IDQuit:
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

var (
	messageStyle = lipgloss.NewStyle().
			Foreground(common.ColorFrostLightBlue). // nord frost blue
			Bold(true)

	messageHintStyle = lipgloss.NewStyle().
				Foreground(common.ColorMutedGray) // nord muted gray
)

func (m Model) viewMessage() string {
	msg := messageStyle.Render(m.message)
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
