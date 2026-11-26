package flag

import (
	"errors"
	"fmt"

	"github.com/natholdallas/natools4go/ask"
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/pwd"
	"github.com/natholdallas/natools4go/rands"

	"webgx/internal/db"
	"webgx/internal/task"
)

func Gensec() error {
	fmt.Println("remake secret script")
	fmt.Println("adm: ", rands.Char(32))
	fmt.Println("usr: ", rands.Char(32))
	fmt.Println("secret remake success, please edit your configuration")
	return nil
}

func RstDB() error {
	fmt.Println("resetting database")
	return db.Reset()
}

func CrtDB() error {
	fmt.Println("creating database")
	return db.AutoCreate()
}

func MigDB() error {
	fmt.Println("migration database")
	return db.Migrate()
}

func SynDB() error {
	fmt.Println("sync database schema")
	return orms.SyncTablesHard(db.Tx, db.Models...)
}

func RstTable() error {
	fmt.Println("reset table structures")
	return orms.ResetTables(db.Tx, db.Models...)
}

func Usr() error {
	fmt.Println("add user script")
	username := ask.Read[string]("username")
	password, err := readPassword()
	if err != nil {
		return err
	}
	if err = checkPassword(password); err != nil {
		return err
	}
	hash, err := pwd.Hash(password)
	if err != nil {
		return err
	}
	v := db.User{Username: username, Password: hash}
	if err := orms.Create(db.Tx, &v); err != nil {
		return err
	}
	fmt.Println("success")
	return nil
}

func Adm() error {
	fmt.Println("add admin script")
	username := ask.Read[string]("username")
	password, err := readPassword()
	if err != nil {
		return err
	}
	if err = checkPassword(password); err != nil {
		return err
	}
	hash, err := pwd.Hash(password)
	if err != nil {
		return err
	}
	v := db.Admin{Username: username, Password: hash}
	if err := orms.Create(db.Tx, &v); err != nil {
		return err
	}
	fmt.Println("success")
	return nil
}

// checkPassword enforces the same 4-72 character bounds the HTTP handlers
// validate, so the CLI cannot create a credential the login endpoint refuses.
func checkPassword(password string) error {
	if n := len([]rune(password)); n < 4 || n > 72 {
		return errors.New("password must be 4-72 characters")
	}
	return nil
}

// readPassword prompts without echoing. ask.Read echoes input, which would print
// the new credential to the terminal and leave it in scrollback.
func readPassword() (string, error) {
	return ask.ReadPassword("password")
}

func Sync() error {
	fmt.Println("sync script")
	return task.Sync()
}

func Mock() error {
	fmt.Println("database mock script")
	return db.Mock()
}
