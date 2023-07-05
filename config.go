package main

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	DevApiKey    string
	SourceFolder string
	Username     string
	Password     string
}

func MakeConfig() error {
	file, err := os.Create(configName)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	file.WriteString(tomlTemplate)
	return nil
}

func ReadConfig() (*Config, error) {
	configfile := configName
	if _, err := os.Stat(configfile); err != nil {
		if err := MakeConfig(); err != nil {
			return nil, err
		}
	}

	var config Config
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
