package conf

import (
	"flag"

	"github.com/natholdallas/natools4go/spew"
)

type FlagConf struct {
	ConfPath  string
	ConfName  string
	AdminAdd  bool
	UserAdd   bool
	RstDB     bool
	Migration bool
	Sync      bool
	Mock      bool
}

func LoadFlag() {
	flag.StringVar(&Flag.ConfPath, "conf", ".", "define config path")
	flag.StringVar(&Flag.ConfName, "confname", "conf", "define config name")
	flag.BoolVar(&Flag.AdminAdd, "adminadd", false, "add admin")
	flag.BoolVar(&Flag.UserAdd, "useradd", false, "add user")
	flag.BoolVar(&Flag.RstDB, "rstdb", false, "reset database, if exists will be clear to default")
	flag.BoolVar(&Flag.Migration, "migration", false, "run migration script")
	flag.BoolVar(&Flag.Sync, "sync", false, "sync data in database")
	flag.BoolVar(&Flag.Mock, "mock", false, "mock data in database")
	flag.Parse()
	spew.JSON(Flag)
}
