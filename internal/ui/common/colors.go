package common

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/rokybeast/zengit/internal/config"
)

var (
	ColorFrostBlue      lipgloss.TerminalColor
	ColorFrostLightBlue lipgloss.TerminalColor
	ColorSnow           lipgloss.TerminalColor
	ColorSnowDark       lipgloss.TerminalColor
	ColorMutedGray      lipgloss.TerminalColor
	ColorMutedGrayDark  lipgloss.TerminalColor
	ColorGreen          lipgloss.TerminalColor
	ColorRed            lipgloss.TerminalColor
	ColorOrange         lipgloss.TerminalColor
	ColorBgDiffAdd      lipgloss.TerminalColor
	ColorBgDiffDelete   lipgloss.TerminalColor
	ColorBgDiffLine     lipgloss.TerminalColor
)

func init() {
	theme := config.GetCurrentTheme()
	ColorFrostBlue = lipgloss.Color(theme.Primary)
	ColorFrostLightBlue = lipgloss.Color(theme.PrimaryLight)
	ColorSnow = lipgloss.Color(theme.Text)
	ColorSnowDark = lipgloss.Color(theme.TextDark)
	ColorMutedGray = lipgloss.Color(theme.Muted)
	ColorMutedGrayDark = lipgloss.Color(theme.MutedDark)
	ColorGreen = lipgloss.Color(theme.Success)
	ColorRed = lipgloss.Color(theme.Error)
	ColorOrange = lipgloss.Color(theme.Warning)
	ColorBgDiffAdd = lipgloss.Color(theme.DiffAddBg)
	ColorBgDiffDelete = lipgloss.Color(theme.DiffDelBg)
	ColorBgDiffLine = lipgloss.Color(theme.DiffLineBg)
}
