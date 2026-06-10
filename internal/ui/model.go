package ui

import (
	"os"

	"gitty/internal/git"
	"gitty/internal/ui/about"
	"gitty/internal/ui/initflow"
	"gitty/internal/ui/menu"
	"gitty/internal/ui/nav"

	tea "github.com/charmbracelet/bubbletea"
)

type state int

const (
	stateNoGit state = iota
	stateGit
	stateInitRepo
	stateNav
	stateAbout
)

type Model struct {
	state    state
	noGit    menu.NoGitModel
	initFlow initflow.Model
	nav      nav.Model
	about    about.Model
	quitting bool
	width    int
	height   int
}

// make a new fresh model and detect the git repo to pick the first state
func New() Model {
	cwd, _ := os.Getwd()
	s := stateNoGit
	if git.IsRepo(cwd) {
		s = stateGit
	}

	return Model{
		state: s,
		noGit: menu.NewNoGit(),
	}
}

func (m Model) Init() tea.Cmd {
	switch m.state {
	case stateNoGit:
		return m.noGit.Init()
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
	case initflow.DoneMsg:
		_ = git.InitRepo(msg.GitIgnore, msg.License)
		m.state = stateGit
		return m, nil
	case nav.PickedMsg:
		// change into the selected repo directory
		_ = os.Chdir(msg.Path)
		m.state = stateGit
		return m, nil
	case nav.BackMsg, about.BackMsg:
		// return to the no-git menu from nav or about
		m.state = stateNoGit
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
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
		return "wow! git repo, main menu coming soon\n"
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
		m.state = stateAbout
		m.about = about.New()
		return m, m.about.Init()
	case "Quit":
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}
