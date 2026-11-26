// Package flag to setup command line flag
package flag

import (
	"github.com/natholdallas/natools4go/flags"

	"webgx/internal/conf"
)

func PreRun() {
	flags.RunE(conf.Flag.Gensec, Gensec)
}

func Run() {
	flags.RunE(conf.Flag.RstDB, RstDB)
	flags.RunE(conf.Flag.CrtDB, CrtDB)
	flags.RunE(conf.Flag.MigDB, MigDB)
	flags.RunE(conf.Flag.SynDB, SynDB)
	flags.RunE(conf.Flag.RstTable, RstTable)
	flags.RunE(conf.Flag.Adm, Adm)
	flags.RunE(conf.Flag.Usr, Usr)
	flags.RunE(conf.Flag.Sync, Sync)
	flags.RunE(conf.Flag.Mock, Mock)
}
