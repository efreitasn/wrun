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
		}),
	)

	set.Add(
		"init",
		"Creates a config file in the current directory",
		cfop.NewCmd(cfop.CmdConfig{
			Fn: cmds.Init,
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
