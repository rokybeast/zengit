package ui

import (
	"fmt"
	"os"

	"gitty/internal/git"
	"gitty/internal/ui/about"
	"gitty/internal/ui/commitflow"
	"gitty/internal/ui/initflow"
	"gitty/internal/ui/menu"
	"gitty/internal/ui/nav"
	"gitty/internal/ui/pushflow"
	"gitty/internal/ui/treeflow"

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
)

type Model struct {
	state      state
	prevState  state // where to go back to from about/sub-screens
	noGit      menu.NoGitModel
	gitMenu    menu.GitModel
	initFlow   initflow.Model
	commitFlow commitflow.Model
	treeFlow   treeflow.Model
	pushFlow   pushflow.Model
	nav        nav.Model
	about      about.Model
	quitting   bool
	width      int
	height     int
	message    string
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
		return m.nav.Init()
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
	case nav.PickedMsg:
		// change into the selected repo directory
		_ = os.Chdir(msg.Path)
		m.state = stateGit
		m.gitMenu = menu.NewGit(m.width, m.height)
		return m, nil
	case nav.BackMsg, about.BackMsg, commitflow.BackMsg, treeflow.BackMsg, pushflow.BackMsg:
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
		updated, cmd = m.nav.Update(msg)
		m.nav = updated.(nav.Model)
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
		return m.nav.View()
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
	case stateMessage:
		return m.viewMessage()
	}

	return ""
}

func (m Model) handleNoGitOption(msg menu.ChoiceMsg) (tea.Model, tea.Cmd) {
	switch msg.Choice {
	case "Initialize a Git Repository":
		m.state = stateInitRepo
		m.initFlow = initflow.New(m.width, m.height)
		return m, m.initFlow.Init()
	case "Navigate to a Git Repository":
		m.state = stateNav
		m.nav = nav.New()
		return m, m.nav.Init()
	case "About gitty":
		m.prevState = stateNoGit
		m.state = stateAbout
		m.about = about.New()
		return m, m.about.Init()
	case "Quit":
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

// handle selections from the git repo menu
func (m Model) handleGitOption(msg menu.GitChoiceMsg) (tea.Model, tea.Cmd) {
	switch msg.Choice {
	case "About gitty":
		m.prevState = stateGit
		m.state = stateAbout
		m.about = about.New()
		return m, m.about.Init()
	case "Commit":
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
		m.state = stateCommit
		m.commitFlow = commitflow.New(m.width, m.height)
		return m, m.commitFlow.Init()
	case "Add Files", "Project Tree":
		if msg.Choice == "Add Files" {
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
		}
		m.prevState = stateGit
		m.state = stateTree
		m.treeFlow = treeflow.New(m.width, m.height, msg.Choice == "Add Files")
		return m, m.treeFlow.Init()
	case "Push Commits":
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
	case "Quit":
		m.quitting = true
		return m, tea.Quit
	}
	// other options will be wired up as we build them
	return m, nil
}

var (
	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#81a1c1")). // nord frost blue
			Bold(true)

	messageHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4c566a")) // nord muted gray
)

func (m Model) viewMessage() string {
	msg := messageStyle.Render(m.message)
	hint := messageHintStyle.Render("\n\npress esc/enter/q to go back")
	fullText := msg + hint
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(fullText)
}
