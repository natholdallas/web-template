package flag

import (
	"fmt"

	"github.com/natholdallas/natools4go/ask"
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/pwd"
	"github.com/natholdallas/natools4go/rands"

	"webtplmst/internal/db"
	"webtplmst/internal/task"
)

func Gensec() {
	fmt.Println("remake secret script")
	fmt.Println("adm: ", rands.Char(32))
	fmt.Println("usr: ", rands.Char(32))
	fmt.Println("secret remake success, please edit your configuration")
}

func RstDB() {
	fmt.Println("resetting database")
	db.Reset()
}

func CrtDB() {
	fmt.Println("creating database")
	db.AutoCreate()
}

func MigDB() {
	fmt.Println("migration database")
	db.Migrate()
}

func SyncDB() {
	fmt.Println("sync database schema")
	db.SyncDB(db.Tx)
}

func RstTable() {
	fmt.Println("reset table structures")
	db.ResetTables(db.Tx)
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
