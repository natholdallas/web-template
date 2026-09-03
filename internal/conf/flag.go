package conf

import (
	"flag"
)

type FlagConf struct {
	ConfPath string
	ConfName string
	ConfType string
	Adm      bool
	Usr      bool
	RstDB    bool
	Migrate  bool
	Sync     bool
	Mock     bool
	Gensec   bool
}

func LoadFlag() {
	flag.StringVar(&Flag.ConfPath, "conf", ".", "config path")
	flag.StringVar(&Flag.ConfName, "confname", "conf", "config name")
	flag.StringVar(&Flag.ConfType, "conftype", "toml", "config type")
	flag.BoolVar(&Flag.Adm, "adm", false, "create admin")
	flag.BoolVar(&Flag.Usr, "usr", false, "create user")
	flag.BoolVar(&Flag.RstDB, "rstdb", false, "reset database to default")
	flag.BoolVar(&Flag.Migrate, "migrate", false, "run migration script")
	flag.BoolVar(&Flag.Sync, "sync", false, "run sync task")
	flag.BoolVar(&Flag.Mock, "mock", false, "create mock data")
	flag.BoolVar(&Flag.Gensec, "gensec", false, "regenerate secrets")
	flag.Parse()
}
