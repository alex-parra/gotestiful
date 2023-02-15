package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".gotestiful"

type config struct {
	Color        bool     `json:"color"`
	Cache        bool     `json:"cache"`
	Cover        bool     `json:"cover"`
	Report       bool     `json:"report"`
	CoverProfile string   `json:"coverProfile"`
	Verbose      bool     `json:"verbose"`
	ListIgnored  bool     `json:"listIgnored"`
	SkipEmpty    bool     `json:"skipEmpty"`
	ListEmpty    bool     `json:"listEmpty"`
	FullCoverage bool     `json:"fullCoverage"`
	Exclude      []string `json:"exclude"`
	TestOutput   string   `json:"testOutput"`
}

// Default config values
var conf = config{
	Color: true,
	Cache: true,
	Cover: true,
	// Report:       false,
	// CoverProfile: "",
	// Verbose:      false,
	// ListIgnored:  false,
	SkipEmpty: true,
	// ListEmpty:    false,
	Exclude: []string{},
}

func GetConfig() (config, error) {
	pwd, err := getPWD()
	if err != nil {
		return config{}, err
	}
	confPath := filepath.Join(pwd, configFileName)

	if fileExists(confPath) {
		confBytes, err := readFile(confPath)
		if err != nil {
			return config{}, fmt.Errorf("failed to read config file: %w", err)
		}

		err = json.Unmarshal(confBytes, &conf)
		if err != nil {
			return config{}, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	return conf, nil
}

// Creates a config file in the current path with default values
func InitConfig() error {
	pwd, err := getPWD()
	if err != nil {
		return err
	}
	confPath := filepath.Join(pwd, configFileName)

	if fileExists(confPath) {
		return fmt.Errorf("config file already exits at %s", confPath)
	}

	data, _ := json.MarshalIndent(conf, "", "  ")

	err = os.WriteFile(confPath, data, 0644)

	if err != nil {
		return fmt.Errorf("failed to init config: %w", err)
	}

	return nil
}
