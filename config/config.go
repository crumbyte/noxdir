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

type Settings struct {
	Path        string   `json:"-"`
	ColorSchema string   `json:"colorSchema"`
	Exclude     []string `json:"exclude"`
	NoEmptyDirs bool     `json:"noEmptyDirs"`
	NoHidden    bool     `json:"noHidden"`
	SimpleColor bool     `json:"simpleColor"`
	UseCache    bool     `json:"useCache"`
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

	if err = json.NewEncoder(settingsFile).Encode(&s); err != nil {
		return nil, err
	}

	if _, err = settingsFile.Seek(0, 0); err != nil {
		return nil, err
	}

	return settingsFile, nil
}
