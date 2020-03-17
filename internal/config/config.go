package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
)

var defaultDelayToKill = 1000
var defaultConfigFilePath = "wrun.json"
var configFileSchemaURL = "https://raw.githubusercontent.com/efreitasn/wrun/master/wrun.schema.json"

// alwaysIgnoreRegExps is a list of regexps that are always ignored.
var alwaysIgnoreRegExps = []*regexp.Regexp{
	regexp.MustCompile("^.git*"),
	regexp.MustCompile(".*wrun.json"),
}

type configFileCmd struct {
	DelayToKill *int     `json:"delayToKill"`
	FatalIfErr  *bool    `json:"fatalIfErr"`
	Terms       []string `json:"terms"`
}

type configFile struct {
	Schema        string          `json:"$schema"`
	DelayToKill   *int            `json:"delayToKill"`
	FatalIfErr    bool            `json:"fatalIfErr"`
	Cmds          []configFileCmd `json:"cmds"`
	IgnoreRegExps []string        `json:"ignoreRegExps"`
}

// Cmd is a command from a config file.
type Cmd struct {
	Terms []string
	// Milliseconds
	DelayToKill int
	FatalIfErr  bool
}

// Config is the data from a config file.
type Config struct {
	Cmds          []Cmd
	IgnoreRegExps []*regexp.Regexp
}

// GetConfig returns the data from the config file.
func GetConfig(configFilePath string) (*Config, error) {
	if configFilePath == "" {
		configFilePath = defaultConfigFilePath
	}

	f, err := os.Open(configFilePath)
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

	c, err := parseConfigFile(cf)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// CreateConfigFile creates a config file in the current directory with default data.
func CreateConfigFile() error {
	file, err := os.OpenFile(
		defaultConfigFilePath,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY,
		0666,
	)
	if err != nil {
		return err
	}

	cmdDefaultFatalIfErr := false
	cf := configFile{
		DelayToKill: &defaultDelayToKill,
		FatalIfErr:  false,
		Cmds: []configFileCmd{configFileCmd{
			Terms:       []string{"echo", "hello", "world"},
			DelayToKill: &defaultDelayToKill,
			FatalIfErr:  &cmdDefaultFatalIfErr,
		}},
		IgnoreRegExps: []string{},
		Schema:        configFileSchemaURL,
	}

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err = enc.Encode(cf); err != nil {
		return err
	}

	return nil
}

// parseConfigFile transforms a configFile to a config.
// Note that this function doesn't perform any kind of validation
// on the configFile.
func parseConfigFile(cf configFile) (*Config, error) {
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

		if len(cfCmd.Terms) == 0 {
			return nil, fmt.Errorf("terms field in cmds[%v] is empty", i)
		}
	}

	globalDelayToKill := defaultDelayToKill
	if cf.DelayToKill != nil {
		globalDelayToKill = *cf.DelayToKill
	}

	globalFatalIfErr := cf.FatalIfErr

	cmds := make([]Cmd, 0, len(cf.Cmds))

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

		cmds = append(cmds, Cmd{
			Terms:       terms,
			DelayToKill: delayToKill,
			FatalIfErr:  fatalIfErr,
		})
	}

	ignoreRegExps := alwaysIgnoreRegExps
	if cf.IgnoreRegExps != nil {
		for _, rxStr := range cf.IgnoreRegExps {
			rx, err := regexp.Compile(rxStr)
			if err != nil {
				return nil, fmt.Errorf("%v regexp is invalid", rxStr)
			}

			ignoreRegExps = append(ignoreRegExps, rx)
		}
	}

	return &Config{
		IgnoreRegExps: ignoreRegExps,
		Cmds:          cmds,
	}, nil
}
