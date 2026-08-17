// Package db to setup database
package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"webtplmst/internal/conf"

	"github.com/fsnotify/fsnotify"
	"github.com/natholdallas/natools4go/orms"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	Tx *gorm.DB
	Tc = context.Background()

	Rx *redis.Client
	Rc = context.Background()

	Dsn string
)

func Connect() {
	switch conf.App.DBDriver {
	case "mysql":
		Dsn = orms.Dsn(conf.App.DBUsername, conf.App.DBPassword, conf.App.DBHost, conf.App.DBPort)
	case "postgres", "postgresql":
		Dsn = fmt.Sprintf("host=%s user=%s password=%s port=%s",
			conf.App.DBHost, conf.App.DBUsername, conf.App.DBPassword, conf.App.DBPort)
	case "sqlite", "sqlite3":
		Dsn = conf.App.DBName
	case "sqlserver", "mssql":
		Dsn = fmt.Sprintf("sqlserver://%s:%s@%s:%s", conf.App.DBUsername, conf.App.DBPassword, conf.App.DBHost, conf.App.DBPort)
	case "clickhouse":
		Dsn = fmt.Sprintf("tcp://%s:%s?username=%s&password=%s", conf.App.DBHost, conf.App.DBPort, conf.App.DBUsername, conf.App.DBPassword)
	default:
		log.Fatalf("unsupported database driver: %s (supported: mysql, postgres, sqlite, sqlserver, clickhouse)", conf.App.DBDriver)
	}

	var dialector gorm.Dialector
	switch conf.App.DBDriver {
	case "mysql":
		if conf.App.DBAutoCreate {
			orms.Prepare(conf.App.DBName, "mysql", Dsn)
		}
		dialector = mysql.Open(Dsn + orms.Queries(conf.App.DBName, conf.App.DBQuery))
	case "postgres", "postgresql":
		if conf.App.DBAutoCreate {
			orms.Prepare(conf.App.DBName, conf.App.DBDriver, Dsn)
		}
		dsn := Dsn + " dbname=" + conf.App.DBName
		if conf.App.DBQuery != "" {
			dsn += " " + conf.App.DBQuery
		}
		dialector = postgres.Open(dsn)
	case "sqlite", "sqlite3":
		dsn := conf.App.DBName
		if conf.App.DBQuery != "" {
			dsn += "?" + conf.App.DBQuery
		}
		dialector = sqlite.Open(dsn)
	case "sqlserver", "mssql":
		if conf.App.DBAutoCreate {
			orms.Prepare(conf.App.DBName, conf.App.DBDriver, Dsn)
		}
		dsn := Dsn + "?database=" + conf.App.DBName
		if conf.App.DBQuery != "" {
			dsn += "&" + conf.App.DBQuery
		}
		dialector = sqlserver.Open(dsn)
	case "clickhouse":
		if conf.App.DBAutoCreate {
			orms.Prepare(conf.App.DBName, conf.App.DBDriver, Dsn)
		}
		dsn := Dsn + "&database=" + conf.App.DBName
		if conf.App.DBQuery != "" {
			dsn += "&" + conf.App.DBQuery
		}
		dialector = clickhouse.Open(dsn)
	default:
		log.Fatalf("unsupported database driver: %s (supported: mysql, postgres, sqlite, sqlserver, clickhouse)", conf.App.DBDriver)
	}

	Tx = orms.New(dialector, &gorm.Config{
		Logger: logger.New(log.New(conf.App.LogWriter(), "[DB] ", log.Ldate), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  conf.App.LogLevelGorm,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		}),
	})

	if conf.App.RedisHost != "" {
		Rx = redis.NewClient(&redis.Options{
			Addr: conf.App.RedisAddr,
			DB:   conf.App.RedisIndex,
		})
	}

	if conf.App.DBAutoMigrate {
		Migration()
	}
}

func Mock() {
	orms.Create(Tx, &Admin{Username: "admin", Password: "123456"})
	orms.Create(Tx, &User{Username: "user", Password: "123456"})
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

func Reload(fsnotify.Event) {
	Tx.Logger.LogMode(conf.App.LogLevelGorm)
}
