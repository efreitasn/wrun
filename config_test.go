package main

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
		res config
	}{
		{
			configFile{
				DelayToKill: &delay700,
				FatalIfErr:  true,
				Cmds: []configFileCmd{
					configFileCmd{
						Terms: []string{"foo", "bar"},
					},
				},
			},
			config{
				cmds: []cmd{
					cmd{
						terms:       []string{"foo", "bar"},
						delayToKill: delay700,
						fatalIfErr:  true,
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
			config{
				cmds: []cmd{
					cmd{
						terms:       []string{},
						delayToKill: delay700,
						fatalIfErr:  true,
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
			config{
				cmds: []cmd{
					cmd{
						terms:       []string{"foo", "bar"},
						delayToKill: delay700,
						fatalIfErr:  boolFalse,
					},
					cmd{
						terms:       []string{"bar", "foo"},
						delayToKill: delay900,
						fatalIfErr:  true,
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
			config{
				cmds: []cmd{
					cmd{
						terms:       []string{"foo", "bar"},
						delayToKill: delay700,
						fatalIfErr:  boolFalse,
					},
					cmd{
						terms:       []string{"bar", "foo"},
						delayToKill: delay0,
						fatalIfErr:  true,
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
			config{
				cmds: []cmd{
					cmd{
						terms:       []string{"foo", "bar"},
						delayToKill: defaultDelayToKill,
						fatalIfErr:  boolFalse,
					},
					cmd{
						terms:       []string{"bar", "foo"},
						delayToKill: defaultDelayToKill,
						fatalIfErr:  true,
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
