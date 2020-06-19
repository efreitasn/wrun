package cmds

import (
	"github.com/efreitasn/cfop"
	"github.com/efreitasn/wrun/v4/internal/config"
	"github.com/efreitasn/wrun/v4/internal/logs"
)

// Init executes the init command.
func Init(cts *cfop.CmdTermsSet) {
	err := config.CreateConfigFile()
	if err != nil {
		logs.Err.Println(err)
	}
}
