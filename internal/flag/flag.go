// Package flag to setup command line flag
package flag

import (
	"fmt"

	"webtplmst/internal/conf"
	"webtplmst/internal/db"
	"webtplmst/internal/task"

	"github.com/natholdallas/natools4go/ask"
	"github.com/natholdallas/natools4go/flags"
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/pwd"
	"github.com/natholdallas/natools4go/rands"
)

func PreRun() {
	flags.Run(conf.Flag.Gensec, Gensec)
}

func Run() {
	flags.Run(conf.Flag.RstDB, RstDB)
	flags.Run(conf.Flag.Migrate, Migrate)
	flags.Run(conf.Flag.Adm, Adm)
	flags.Run(conf.Flag.Usr, Usr)
	flags.Run(conf.Flag.Sync, Sync)
	flags.Run(conf.Flag.Mock, Mock)
}

func RstDB() {
	fmt.Println("resetting database")
	db.Reset()
}

func CrtDB() {
	fmt.Println("creating database")
	db.AutoCreate()
}

func Migrate() {
	fmt.Println("migration database")
	db.Migrate()
}

func Usr() {
	fmt.Println("add user script")
	username := ask.Read[string]("username")
	password := ask.Read[string]("password")
	hash, err := pwd.Hash(password)
	if err != nil {
		fmt.Println(err)
		return
	}
	v := db.User{Username: username, Password: hash}
	if err := orms.Create(db.Tx, &v); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("success")
	}
}

func Adm() {
	fmt.Println("add admin script")
	username := ask.Read[string]("username")
	password := ask.Read[string]("password")
	hash, err := pwd.Hash(password)
	if err != nil {
		fmt.Println(err)
		return
	}
	v := db.Admin{Username: username, Password: hash}
	if err := orms.Create(db.Tx, &v); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("success")
	}
}

func Sync() {
	fmt.Println("sync script")
	task.Sync()
}

func Mock() {
	fmt.Println("database mock script")
	db.Mock()
}

func Gensec() {
	fmt.Println("remake secret script")
	fmt.Println("adm: ", rands.Char(32))
	fmt.Println("usr: ", rands.Char(32))
	fmt.Println("secret remake success, please edit your configuration")
}
