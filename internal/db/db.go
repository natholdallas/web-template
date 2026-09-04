// Package db to setup database
package db

import (
	"context"

	"github.com/fsnotify/fsnotify"
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/pwd"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"webtplmst/internal/conf"
)

var (
	Tx *gorm.DB
	Tc = context.Background()
	Rx *redis.Client
	Rc = context.Background()
)

func Connect() {
	Tx = orms.MustNew(orms.MustDialector(
		conf.App.DBDriver,
		conf.App.DBDsn,
		conf.App.DBName,
		conf.App.DBQuery,
		conf.App.DBAutoCreate,
	), &gorm.Config{
		Logger: orms.LogPreset(conf.App.LogWriter(), conf.App.LogLevelGorm),
	})
	Rx = redis.NewClient(&redis.Options{
		Addr: conf.App.RedisAddr,
		DB:   conf.App.RedisIndex,
	})
	if conf.App.DBAutoMigrate {
		Migrate()
	}
}

func Mock() {
	orms.Create(Tx, &Admin{Username: "admin", Password: pwd.TryHash("123456")})
	orms.Create(Tx, &User{Username: "user", Password: pwd.TryHash("123456")})
}

var Models = []any{&Admin{}, &User{}, &Rate{}, &Media{}, &SysConf{}}

func Migrate() {
	orms.MustAutoMigrate(Tx, Models...)
}

func Reset() {
	orms.Reset(conf.App.DBName, conf.App.DBDriver, conf.App.DBDsn)
}

func AutoCreate() {
	orms.AutoCreate(conf.App.DBName, conf.App.DBDriver, conf.App.DBDsn)
}

func Reload(fsnotify.Event) {
	Tx.Logger.LogMode(conf.App.LogLevelGorm)
}
