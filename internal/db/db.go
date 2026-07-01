// Package db to setup database
package db

import (
	"context"
	"log"
	"time"

	"webtplmst/internal/conf"

	"github.com/fsnotify/fsnotify"
	"github.com/natholdallas/natools4go/orms"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	Tx *gorm.DB
	Tc context.Context = context.TODO()
	Tl logger.Config   = logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  conf.App.LogLevelGorm,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	}

	Rx *redis.Client
	Rc context.Context = context.TODO()

	Dsn     string = orms.Dsn(conf.App.DBUsername, conf.App.DBPassword, conf.App.DBHost, conf.App.DBPort)
	Queries string = orms.Queries(conf.App.DBName, conf.App.DBQuery)
)

// autocreate database
func init() {
	orms.EnsureDB(conf.App.DBName, "mysql", Dsn)
}

// database context initialize
func init() {
	Tx = orms.New(mysql.Open(Dsn+Queries), &gorm.Config{
		Logger: logger.New(log.New(conf.App.LogWriter(), "[DB] ", log.Ldate), Tl),
	})
	Rx = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   10,
	})
}

func Mock() {
	orms.Create(Tx, &Admin{Username: "admin", Password: "123456"})
	orms.Create(Tx, &User{Username: "user", Password: "123456"})
}

func Migration() {
	Tx.
		Set("gorm:table_options", "COLLATE=utf8mb4_bin").
		AutoMigrate(
			&Admin{},
			&User{},
			&Rate{},
			&Media{},
		)
}

func Reload(fsnotify.Event) {
	Tx.Logger.LogMode(conf.App.LogLevelGorm)
}
