package main

import (
	"os"

	"github.com/efreitasn/cfop"
	"github.com/efreitasn/wrun/v4/cmd/wrun/internal/cmds"
	"github.com/efreitasn/wrun/v4/internal/logs"
)

func main() {
	if err := startCmd(os.Args); err != nil {
		logs.Err.Println(err)
	}
}

func startCmd(args []string) error {
	set := cfop.NewSubcmdsSet()

	set.Add(
		"start",
		"Starts watching files in the current directory.",
		cfop.NewCmd(cfop.CmdConfig{
			Fn: cmds.Start,
			Options: []cfop.CmdOption{
				cfop.CmdOption{
					T:           cfop.TermString,
					Name:        "file",
					Alias:       "f",
					Description: "path for the config file",
				},
			},
			Flags: []cfop.CmdFlag{
				cfop.CmdFlag{
					Name:        "no-events",
					Alias:       "ne",
					Description: "whether to log events",
				},
				cfop.CmdFlag{
					Name:        "quiet",
					Alias:       "q",
					Description: "whether to log anything at all",
				},
			},
		}),
	)

	set.Add(
		"init",
		"Creates a config file in the current directory",
		cfop.NewCmd(cfop.CmdConfig{
			Fn: cmds.Init,
		}),
	)

	set.Add(
		"version",
		"Prints the version",
		cfop.NewCmd(cfop.CmdConfig{
			Fn: cmds.Version,
		}),
	)

	err := cfop.Init(
		"wrun",
		"Run commands whenever the contents in the current directory change",
		args,
		set,
	)
	if err != nil {
		return err
	}

	return nil
}
