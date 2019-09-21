package main

import (
	"encoding/json"
	"errors"
	"os"
)

// Config is the config file's structure.
type Config struct {
	Cmd    []string `json:"CMD"`
	PreCmd []string `json:"PRECMD"`
}

func getConfig() (*Config, error) {
	f, err := os.Open("wrun.json")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("wrun.json config file doesn't exist")
		}

		return nil, err
	}
	defer f.Close()

	var config Config

	jsonDec := json.NewDecoder(f)

	err = jsonDec.Decode(&config)
	if err != nil {
		return nil, err
	}

	if len(config.Cmd) == 0 {
		return nil, errors.New("CMD field in wrun.json is empty")
	}

	return &config, nil
}
