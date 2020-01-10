package config

import (
	"reflect"
	"regexp"
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
		err error
	}{
		{
			configFile{
				DelayToKill:   &delay700,
				FatalIfErr:    true,
				IgnoreRegExps: []string{"aa.*"},
				Cmds: []configFileCmd{
					configFileCmd{
						Terms: []string{"foo", "bar"},
					},
				},
			},
			Config{
				IgnoreRegExps: append(alwaysIgnoreRegExps, regexp.MustCompile("aa.*")),
				Cmds: []Cmd{
					Cmd{
						Terms:       []string{"foo", "bar"},
						DelayToKill: delay700,
						FatalIfErr:  true,
					},
				},
			},
			nil,
		},
		{
			configFile{
				DelayToKill: &delay700,
				FatalIfErr:  true,
				Cmds: []configFileCmd{
					configFileCmd{
						Terms: []string{"echo", "a"},
					},
				},
			},
			Config{
				IgnoreRegExps: alwaysIgnoreRegExps,
				Cmds: []Cmd{
					Cmd{
						Terms:       []string{"echo", "a"},
						DelayToKill: delay700,
						FatalIfErr:  true,
					},
				},
			},
			nil,
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
				IgnoreRegExps: alwaysIgnoreRegExps,
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
			nil,
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
				IgnoreRegExps: alwaysIgnoreRegExps,
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
			nil,
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
				IgnoreRegExps: alwaysIgnoreRegExps,
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
			nil,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			res, err := parseConfigFile(test.cf)

			if err != test.err {
				t.Fatalf("got %v, want %v", err, test.err)
			}

			if !reflect.DeepEqual(res.Cmds, test.res.Cmds) {
				t.Errorf("got %v, want %v", res.Cmds, test.res.Cmds)
			}

			resRegExpsStr := make([]string, 0)
			expectedRegExpsStr := make([]string, 0)

			if res.IgnoreRegExps != nil {
				for _, rx := range res.IgnoreRegExps {
					resRegExpsStr = append(resRegExpsStr, rx.String())
				}
			}

			if test.res.IgnoreRegExps != nil {
				for _, rx := range test.res.IgnoreRegExps {
					expectedRegExpsStr = append(expectedRegExpsStr, rx.String())
				}
			}

			if !reflect.DeepEqual(resRegExpsStr, expectedRegExpsStr) {
				t.Errorf("got %v, want %v", resRegExpsStr, expectedRegExpsStr)
			}
		})
	}
}
