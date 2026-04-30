// Package db to setup database
package db

import (
	"context"
	"log"
	"os"
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

	Rx *redis.Client
	Rc context.Context = context.TODO()

	Dsn     string = orms.Dsn(conf.App.DBUsername, conf.App.DBPassword, conf.App.DBHost, conf.App.DBPort)
	Queries string = orms.Queries(conf.App.DBName, conf.App.DBQuery)
)

// database context initialize
func init() {
	orms.EnsureDB(conf.App.DBName, "mysql", Dsn)
	Tx = orms.New(mysql.Open(Dsn+Queries), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond, // Slow SQL threshold
			LogLevel:                  conf.App.LogLevelGorm,  // Log level
			IgnoreRecordNotFoundError: true,                   // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,                   // Colorful Logging
		}),
	})
	Migration()
}

func Mock() {
}

func Migration() {
	Tx.
		Set("gorm:table_options", "COLLATE=utf8mb4_bin").
		AutoMigrate(
			&Admin{},
			&User{},
			// &Rate{},
			// &Media{},
		)
}

func Reload(fsnotify.Event) {
	Tx.Logger.LogMode(conf.App.LogLevelGorm)
}
