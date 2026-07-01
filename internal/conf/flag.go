package conf

import (
	"flag"

	"github.com/natholdallas/natools4go/spew"
)

type FlagConf struct {
	ConfPath     string
	ConfName     string
	ConfType     string
	Adm          bool
	Usr          bool
	RstDB        bool
	Migration    bool
	Sync         bool
	Mock         bool
	RemakeSecret bool
}

func LoadFlag() {
	flag.StringVar(&Flag.ConfPath, "conf", ".", "define config path")
	flag.StringVar(&Flag.ConfName, "confname", "conf", "define config name")
	flag.StringVar(&Flag.ConfType, "conftype", "toml", "define config type")
	flag.BoolVar(&Flag.Adm, "adm", false, "add admin")
	flag.BoolVar(&Flag.Usr, "usr", false, "add user")
	flag.BoolVar(&Flag.RstDB, "rstdb", false, "reset database, if exists will be clear to default")
	flag.BoolVar(&Flag.Migration, "migration", false, "run migration script")
	flag.BoolVar(&Flag.Sync, "sync", false, "sync data in database")
	flag.BoolVar(&Flag.Mock, "mock", false, "mock data in database")
	flag.BoolVar(&Flag.RemakeSecret, "remake-secret", false, "remake secret in toml configuration")
	flag.Parse()
	spew.JSON(Flag)
}
