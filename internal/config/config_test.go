package config

import (
	"reflect"
	"strconv"
	"testing"
)

func TestParseConfigFile(t *testing.T) {
	delay0 := 0
	delay700 := 700
	delay900 := 900
	boolFalse := false

	tests := []struct {
		cf  configFile
		res Config
	}{
		{
			configFile{
				DelayToKill: &delay700,
				FatalIfErr:  true,
				IgnoreGlobs: []string{"aa*"},
				Cmds: []configFileCmd{
					configFileCmd{
						Terms: []string{"foo", "bar"},
					},
				},
			},
			Config{
				IgnoreGlobs: append(alwaysIgnoreGlobs, "aa*"),
				Cmds: []Cmd{
					Cmd{
						Terms:       []string{"foo", "bar"},
						DelayToKill: delay700,
						FatalIfErr:  true,
					},
				},
			},
		},
		{
			configFile{
				DelayToKill: &delay700,
				FatalIfErr:  true,
				Cmds: []configFileCmd{
					configFileCmd{
						Terms: nil,
					},
				},
			},
			Config{
				IgnoreGlobs: alwaysIgnoreGlobs,
				Cmds: []Cmd{
					Cmd{
						Terms:       []string{},
						DelayToKill: delay700,
						FatalIfErr:  true,
					},
				},
			},
		},
		{
			configFile{
				DelayToKill: &delay900,
				FatalIfErr:  true,
				Cmds: []configFileCmd{
					configFileCmd{
						FatalIfErr:  &boolFalse,
						DelayToKill: &delay700,
						Terms:       []string{"foo", "bar"},
					},
					configFileCmd{
						Terms: []string{"bar", "foo"},
					},
				},
			},
			Config{
				IgnoreGlobs: alwaysIgnoreGlobs,
				Cmds: []Cmd{
					Cmd{
						Terms:       []string{"foo", "bar"},
						DelayToKill: delay700,
						FatalIfErr:  boolFalse,
					},
					Cmd{
						Terms:       []string{"bar", "foo"},
						DelayToKill: delay900,
						FatalIfErr:  true,
					},
				},
			},
		},
		{
			configFile{
				DelayToKill: &delay900,
				FatalIfErr:  true,
				Cmds: []configFileCmd{
					configFileCmd{
						FatalIfErr:  &boolFalse,
						DelayToKill: &delay700,
						Terms:       []string{"foo", "bar"},
					},
					configFileCmd{
						DelayToKill: &delay0,
						Terms:       []string{"bar", "foo"},
					},
				},
			},
			Config{
				IgnoreGlobs: alwaysIgnoreGlobs,
				Cmds: []Cmd{
					Cmd{
						Terms:       []string{"foo", "bar"},
						DelayToKill: delay700,
						FatalIfErr:  boolFalse,
					},
					Cmd{
						Terms:       []string{"bar", "foo"},
						DelayToKill: delay0,
						FatalIfErr:  true,
					},
				},
			},
		},
		{
			configFile{
				FatalIfErr: true,
				Cmds: []configFileCmd{
					configFileCmd{
						FatalIfErr: &boolFalse,
						Terms:      []string{"foo", "bar"},
					},
					configFileCmd{
						Terms: []string{"bar", "foo"},
					},
				},
			},
			Config{
				IgnoreGlobs: alwaysIgnoreGlobs,
				Cmds: []Cmd{
					Cmd{
						Terms:       []string{"foo", "bar"},
						DelayToKill: defaultDelayToKill,
						FatalIfErr:  boolFalse,
					},
					Cmd{
						Terms:       []string{"bar", "foo"},
						DelayToKill: defaultDelayToKill,
						FatalIfErr:  true,
					},
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			res := parseConfigFile(test.cf)

			if !reflect.DeepEqual(res, test.res) {
				t.Errorf("got %v, want %v", res, test.res)
			}
		})
	}
}
