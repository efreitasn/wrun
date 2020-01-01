package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const defaultDelayToKill = 1000
const configFileName = "wrun.json"

type configFileCmd struct {
	DelayToKill *int     `json:"delayToKill"`
	FatalIfErr  *bool    `json:"fatalIfErr"`
	Terms       []string `json:"terms"`
}

type configFile struct {
	DelayToKill *int            `json:"delayToKill"`
	FatalIfErr  bool            `json:"fatalIfErr"`
	Cmds        []configFileCmd `json:"cmds"`
}

type cmd struct {
	terms []string
	// Milliseconds
	delayToKill int
	fatalIfErr  bool
}

type config struct {
	cmds []cmd
}

func getConfig() (*config, error) {
	f, err := os.Open(configFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("file doesn't exist")
		}

		return nil, err
	}
	defer f.Close()

	var cf configFile

	jsonDec := json.NewDecoder(f)

	err = jsonDec.Decode(&cf)
	if err != nil {
		return nil, err
	}

	if cf.Cmds == nil {
		return nil, errors.New("missing cmds field")
	}

	if len(cf.Cmds) == 0 {
		return nil, errors.New("cmds field is empty")
	}

	for i, cfCmd := range cf.Cmds {
		if cfCmd.Terms == nil {
			return nil, fmt.Errorf("missing terms field in cmds[%v]", i)
		}
	}

	c := parseConfigFile(cf)

	return &c, nil
}

// parseConfigFile transforms a configFile to a config.
// Note that this function doesn't perform any kind of validation
// on the configFile.
func parseConfigFile(cf configFile) config {
	globalDelayToKill := defaultDelayToKill
	if cf.DelayToKill != nil {
		globalDelayToKill = *cf.DelayToKill
	}

	globalFatalIfErr := cf.FatalIfErr

	cmds := make([]cmd, 0, len(cf.Cmds))

	for _, configCmd := range cf.Cmds {
		delayToKill := globalDelayToKill
		if configCmd.DelayToKill != nil {
			delayToKill = *configCmd.DelayToKill
		}

		fatalIfErr := globalFatalIfErr
		if configCmd.FatalIfErr != nil {
			fatalIfErr = *configCmd.FatalIfErr
		}

		var terms []string

		if configCmd.Terms != nil {
			terms = configCmd.Terms
		} else {
			terms = make([]string, 0)
		}

		cmds = append(cmds, cmd{
			terms:       terms,
			delayToKill: delayToKill,
			fatalIfErr:  fatalIfErr,
		})
	}

	return config{
		cmds: cmds,
	}
}
