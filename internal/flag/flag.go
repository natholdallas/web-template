// Package flag to setup command line flag
package flag

import (
	"webtplmst/internal/conf"

	"github.com/natholdallas/natools4go/flags"
)

func PreRun() {
	flags.Run(conf.Flag.Gensec, Gensec)
}

func Run() {
	flags.Run(conf.Flag.RstDB, RstDB)
	flags.Run(conf.Flag.CrtDB, CrtDB)
	flags.Run(conf.Flag.MigDB, MigDB)
	flags.Run(conf.Flag.Adm, Adm)
	flags.Run(conf.Flag.Usr, Usr)
	flags.Run(conf.Flag.Sync, Sync)
	flags.Run(conf.Flag.Mock, Mock)
}
