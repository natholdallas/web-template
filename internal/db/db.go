// Package db to setup database
package db

import (
	"context"
	"webtplmst/internal/conf"

	"github.com/fsnotify/fsnotify"
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/redisx"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	Tx *gorm.DB
	Tc = context.Background()

	Rx *redis.Client
	Rc = context.Background()
)

func Connect() {
	dsn := orms.MustDsn(conf.App.DBDriver, conf.App.DBName, conf.App.DBUsername, conf.App.DBPassword, conf.App.DBHost, conf.App.DBPort)
	dialector := orms.MustDialector(conf.App.DBDriver, dsn, conf.App.DBName, conf.App.DBQuery, conf.App.DBAutoCreate)
	Tx = orms.MustNew(dialector, &gorm.Config{
		Logger: orms.LogPreset(conf.App.LogWriter(), conf.App.LogLevelGorm),
	})
	Rx = redisx.New(&redis.Options{
		Addr: conf.App.RedisAddr,
		DB:   conf.App.RedisIndex,
	})
	AutoMigration()
}

func Mock() {
	orms.Create(Tx, &Admin{Username: "admin", Password: "123456"})
	orms.Create(Tx, &User{Username: "user", Password: "123456"})
}

func AutoMigration() {
	if conf.App.DBAutoMigrate {
		Migration()
	}
}

func Migration() {
	tx := Tx
	switch conf.App.DBDriver {
	case "mysql":
		tx = tx.Set("gorm:table_options", "COLLATE=utf8mb4_bin")
	case "sqlserver", "mssql":
		tx = tx.Set("gorm:table_options", "COLLATE=SQL_Latin1_General_CP1_CS_AS")
	}
	tx.AutoMigrate(
		&Admin{},
		&User{},
		&Rate{},
		&Media{},
	)
}

func ResetDB() {
	dsn := orms.MustDsn(conf.App.DBDriver, conf.App.DBName, conf.App.DBUsername, conf.App.DBPassword, conf.App.DBHost, conf.App.DBPort)
	orms.Reset(conf.App.DBName, conf.App.DBDriver, dsn)
}

func AutoCreateDB() {
	dsn := orms.MustDsn(conf.App.DBDriver, conf.App.DBName, conf.App.DBUsername, conf.App.DBPassword, conf.App.DBHost, conf.App.DBPort)
	orms.AutoCreate(conf.App.DBName, conf.App.DBDriver, dsn)
}

func Reload(fsnotify.Event) {
	Tx.Logger.LogMode(conf.App.LogLevelGorm)
}
