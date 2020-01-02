package main

import (
	"os"

	"github.com/efreitasn/cfop"
	"github.com/efreitasn/wrun/cmd/wrun/internal/cmds"
	"github.com/efreitasn/wrun/internal/logs"
)

func main() {
	set := cfop.NewSubcmdsSet()

	set.Add(
		"start",
		"Starts watching files in the current directory.",
		cfop.NewCmd(cfop.CmdConfig{
			Fn: cmds.Start,
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
		"Run commands whenever files change",
		os.Args,
		set,
	)
	if err != nil {
		logs.Err.Println(err)
	}
}
