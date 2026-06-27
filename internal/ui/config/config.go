package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

//go:embed config.toml
var defaultConfigFile []byte

type Config struct {
	Config  CoreConfig       `toml:"config"`
	Themes  map[string]Theme `toml:"themes"`
	Commits CommitConfig     `toml:"commits"`
}

type CoreConfig struct {
	Theme string `toml:"theme"`
}

type Theme struct {
	Primary      string `toml:"primary"`
	PrimaryLight string `toml:"primary_light"`
	Text         string `toml:"text"`
	TextDark     string `toml:"text_dark"`
	Muted        string `toml:"muted"`
	MutedDark    string `toml:"muted_dark"`
	Success      string `toml:"success"`
	Error        string `toml:"error"`
	Warning      string `toml:"warning"`
	DiffAddBg    string `toml:"diff_add_bg"`
	DiffDelBg    string `toml:"diff_del_bg"`
	DiffLineBg   string `toml:"diff_line_bg"`
}

type CommitConfig struct {
	Templates bool          `toml:"templates"`
	Entries   []CommitEntry `toml:"entries"`
}

type CommitEntry struct {
	Name        string `toml:"name"`
	Description string `toml:"description"`
}

var (
	AppConfig Config
	ConfigErr error
)

func LoadConfig() error {
	// load defaults
	if _, err := toml.Decode(string(defaultConfigFile), &AppConfig); err != nil {
		return fmt.Errorf("failed to parse default config: %v", err)
	}

	// try to load user config if exists
	home, err := os.UserHomeDir()
	if err == nil {
		userConfigPath := filepath.Join(home, ".config", "gitty", "config.toml") // no windows support for now..
		if data, err := os.ReadFile(userConfigPath); err == nil {
			// override with user config
			if _, err := toml.Decode(string(data), &AppConfig); err != nil {
				return fmt.Errorf("failed to parse user config at %s: %v", userConfigPath, err)
			}
		}
	}
	return nil
}

func GetCurrentTheme() Theme {
	if AppConfig.Themes == nil {
		_ = LoadConfig()
	}
	if theme, ok := AppConfig.Themes[AppConfig.Config.Theme]; ok {
		return theme
	}
	// just in case, if the themes dont work, a fallback
	return AppConfig.Themes["nord"]
}

func init() {
	ConfigErr = LoadConfig()
}
