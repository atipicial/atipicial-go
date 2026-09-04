package app

import (
	"fmt"
	"os"
	"runtime"

	"github.com/atipicial/atipicial-go/cli/query"
	"github.com/atipicial/atipicial-go/cli/server"
	"github.com/atipicial/atipicial-go/cli/smartcontract"
	"github.com/atipicial/atipicial-go/cli/util"
	"github.com/atipicial/atipicial-go/cli/vm"
	"github.com/atipicial/atipicial-go/cli/wallet"
	"github.com/atipicial/atipicial-go/pkg/config"
	"github.com/urfave/cli/v2"
)

func versionPrinter(c *cli.Context) {
	_, _ = fmt.Fprintf(c.App.Writer, "AtipicialGo\nVersion: %s\nGoVersion: %s\n",
		config.Version,
		runtime.Version(),
	)
}

// New creates a AtipicialGo instance of [cli.App] with all commands included.
func New() *cli.App {
	cli.VersionPrinter = versionPrinter
	ctl := cli.NewApp()
	ctl.Name = "atipicial-go"
	ctl.Version = config.Version
	ctl.Usage = "Official Go client for Atipicial"
	ctl.ErrWriter = os.Stdout

	ctl.Commands = append(ctl.Commands, server.NewCommands()...)
	ctl.Commands = append(ctl.Commands, smartcontract.NewCommands()...)
	ctl.Commands = append(ctl.Commands, wallet.NewCommands()...)
	ctl.Commands = append(ctl.Commands, vm.NewCommands()...)
	ctl.Commands = append(ctl.Commands, util.NewCommands()...)
	ctl.Commands = append(ctl.Commands, query.NewCommands()...)
	return ctl
}
