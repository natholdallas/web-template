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
	CrtDB    bool
	MigDB    bool
	SynDB    bool
	RstTable bool
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
	flag.BoolVar(&Flag.RstDB, "db-reset", false, "reset database to default")
	flag.BoolVar(&Flag.CrtDB, "db-create", false, "create database if not exists")
	flag.BoolVar(&Flag.MigDB, "db-migrate", false, "run migration script")
	flag.BoolVar(&Flag.SynDB, "db-sync", false, "sync table schemas with models (add/drop/reorder, keeps data)")
	flag.BoolVar(&Flag.RstTable, "db-reset-table", false, "reset table structures to match models (drops data)")
	flag.BoolVar(&Flag.Sync, "sync", false, "run sync task")
	flag.BoolVar(&Flag.Mock, "mock", false, "create mock data")
	flag.BoolVar(&Flag.Gensec, "gensec", false, "regenerate secrets")
	flag.Parse()
}
