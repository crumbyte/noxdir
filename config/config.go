package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	DirName  = ".noxdir"
	FileName = "settings.json"
)

type DriveBindings struct {
	LevelDown []string `json:"levelDown"`
}

type DirBindings struct {
	LevelUp         []string `json:"levelUp"`
	LevelDown       []string `json:"levelDown"`
	Delete          []string `json:"delete"`
	TopFiles        []string `json:"topFiles"`
	TopDirs         []string `json:"topDirs"`
	FilesOnly       []string `json:"filesOnly"`
	DirsOnly        []string `json:"dirsOnly"`
	NameFilter      []string `json:"nameFilter"`
	ToggleSelectAll []string `json:"toggleSelectAll"`
	Chart           []string `json:"chart"`
	Diff            []string `json:"diff"`
}

type Bindings struct {
	DriveBindings DriveBindings `json:"driveBindings"`
	DirBindings   DirBindings   `json:"dirBindings"`
	Explore       []string      `json:"explore"`
	Quit          []string      `json:"quit"`
	Refresh       []string      `json:"refresh"`
	Help          []string      `json:"help"`
	Config        []string      `json:"config"`
}

type Settings struct {
	Path        string   `json:"-"`
	ColorSchema string   `json:"colorSchema"`
	Exclude     []string `json:"exclude"`
	NoEmptyDirs bool     `json:"noEmptyDirs"`
	NoHidden    bool     `json:"noHidden"`
	SimpleColor bool     `json:"simpleColor"`
	UseCache    bool     `json:"useCache"`
	Bindings    Bindings `json:"bindings"`
}

func LoadSettings() (*Settings, error) {
	var s Settings

	cfgPath, err := ResolveConfigPath(DirName)
	if err != nil {
		return nil, err
	}

	settingsFile, err := openSettings(filepath.Join(cfgPath, FileName))
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = settingsFile.Close()
	}()

	if err = json.NewDecoder(settingsFile).Decode(&s); err != nil {
		return nil, fmt.Errorf("cannot parse settings file: %w", err)
	}

	s.Path = cfgPath

	return &s, nil
}

func (s Settings) ConfigPath() string {
	return filepath.Join(s.Path, FileName)
}

func ResolveConfigPath(configDir string) (string, error) {
	configPath, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve config path: %w", err)
	}

	if runtime.GOOS == "windows" {
		configPath = os.Getenv("LocalAppData")

		if len(configPath) == 0 {
			return "", errors.New("local app data folder not found")
		}
	}

	fullPath := filepath.Join(configPath, configDir)

	if err = os.MkdirAll(fullPath, 0750); err != nil {
		return "", fmt.Errorf("create config path: %w", err)
	}

	return fullPath, nil
}

func openSettings(settingsPath string) (*os.File, error) {
	var s Settings

	settingsFile, err := os.Open(settingsPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("cannot open settings file: %w", err)
	}

	if err == nil {
		return settingsFile, nil
	}

	settingsFile, err = os.Create(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("create default settings: %w", err)
	}

	encoder := json.NewEncoder(settingsFile)
	encoder.SetIndent("", "  ")

	if err = encoder.Encode(&s); err != nil {
		return nil, err
	}

	if _, err = settingsFile.Seek(0, 0); err != nil {
		return nil, err
	}

	return settingsFile, nil
}
