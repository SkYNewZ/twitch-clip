package config

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

const (
	configFileName      = "config.yaml"
	configDirectoryName = "Twitch Clip"
)

type Config struct {
	Notifications []string `json:"notifications,omitempty" yaml:"notifications,flow"`
}

func defaultConfig() *Config {
	return &Config{}
}

// Parse load config stored on disk
func Parse() *Config {
	var config = defaultConfig()

	dir, err := os.UserConfigDir()
	if err != nil {
		log.Errorf("cannot load config: %v", err)
		return config
	}

	var file = filepath.Join(dir, configDirectoryName, configFileName)
	log.Printf("using config file: %s", file)

	data, err := os.ReadFile(file)
	if err != nil {
		log.Errorf("cannot load config: %v", err)
		return config
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		log.Errorf("cannot load config: %v", err)
		return config
	}

	log.Printf("%d notification(s) has been configured", len(config.Notifications))
	return config
}
