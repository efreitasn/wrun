package cmds

import (
	"fmt"

	"github.com/efreitasn/cfop"
)

var version = "dev"

// Version executes the version command.
func Version(cts *cfop.CmdTermsSet) {
	fmt.Println(version)
}
