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
	"github.com/natholdallas/natools4go/rands"
	"github.com/spf13/viper"
)

func Setup() {
	flags.Run(conf.Flag.RstDB, RstDB)
	flags.Run(conf.Flag.Migration, Migration)
	flags.Run(conf.Flag.Adm, Adm)
	flags.Run(conf.Flag.Usr, Usr)
	flags.Run(conf.Flag.Sync, Sync)
	flags.Run(conf.Flag.Mock, Mock)
	flags.Run(conf.Flag.RemakeSecret, RemakeSecret)
}

func RstDB() {
	fmt.Println("resetting database")
	orms.ResetDB(conf.App.DBName, "mysql", db.Dsn)
}

func Migration() {
	fmt.Println("migration database")
	db.Migration()
}

func Usr() {
	fmt.Println("add user script")
	username := ask.Read[string]("username")
	password := ask.Read[string]("password")
	v := db.User{Username: username, Password: password}
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
	v := db.Admin{Username: username, Password: password}
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

func RemakeSecret() {
	fmt.Println("remake secret script")
	viper.Set("secret.adm", rands.Char(32))
	viper.Set("secret.usr", rands.Char(32))
	if err := viper.WriteConfig(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("secret remake success")
	}
}
