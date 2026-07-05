// SPDX-License-Identifier: GPL-3.0-or-later

package gtk

import (
	"fmt"
	"os"

	appconfig "github.com/alyraffauf/switchyard/internal/config"
	"github.com/alyraffauf/switchyard/internal/routing"
)

type Config = appconfig.Config
type Redirection = routing.Redirection
type Condition = routing.Condition
type Rule = routing.Rule

func configPath() string {
	return appconfig.Path()
}

func loadConfig() *Config {
	config, err := appconfig.Load(configPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to parse config file: %v\n", err)
		fmt.Fprintf(os.Stderr, "Using default configuration\n")
	}
	return config
}

func saveConfig(config *Config) error {
	return appconfig.Save(configPath(), config)
}

func exportConfig(config *Config, path string) error {
	return appconfig.Export(path, config)
}

func importConfig(config *Config, path string) error {
	importedConfig, err := appconfig.Import(path)
	if err != nil {
		return err
	}

	*config = *importedConfig
	return saveConfig(config)
}
