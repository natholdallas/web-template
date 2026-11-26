// Package db to setup database
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/pwd"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"webgx/internal/conf"
)

var (
	Tx *gorm.DB
	Rx *redis.Client
	Rc context.Context
)

var Models = []any{
	&Admin{},
	&User{},
	&Rate{},
	&Media{},
	&SysConf{},
}

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
	configurePool()
	Rx = redis.NewClient(&redis.Options{
		Addr: conf.App.RedisAddr,
		DB:   conf.App.RedisIndex,
	})
	Rc = context.Background()
	// Redis backs refresh-token storage, so an unreachable server must fail the
	// boot instead of surfacing on the first sign-in attempt.
	if err := Rx.Ping(Rc).Err(); err != nil {
		panic(fmt.Errorf("connect redis: %w", err))
	}
	if conf.App.DBAutoMigrate {
		if err := Migrate(); err != nil {
			panic(fmt.Errorf("migrate database: %w", err))
		}
	}
}

// configurePool bounds the connection pool. SQLite takes a single writer on
// purpose: concurrent writers on one file surface as "database is locked"
// instead of waiting. Server databases get a normal pool with short-lived idle
// connections so a restarted database server does not leave dead sockets.
func configurePool() {
	sqlDB, err := Tx.DB()
	if err != nil {
		panic(fmt.Errorf("get database handle: %w", err))
	}
	// orms.New already pins file-backed SQLite to one connection (concurrent
	// writers on one file surface as "database is locked").
	if conf.App.DBDriver == "sqlite" || conf.App.DBDriver == "sqlite3" {
		return
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)
}

// Close releases the database and Redis handles. It is safe to call once at
// process shutdown after the HTTP server has stopped.
func Close() {
	if Tx != nil {
		if sqlDB, err := Tx.DB(); err != nil {
			log.Warn("get database handle: ", err)
		} else if err := sqlDB.Close(); err != nil {
			log.Warn("close database: ", err)
		}
	}
	if Rx != nil {
		if err := Rx.Close(); err != nil {
			log.Warn("close redis: ", err)
		}
	}
}

// Mock inserts the demo admin and user. It stops at the first failure so a
// duplicate seed (unique username) is reported instead of silently ignored.
func Mock() error {
	if err := orms.Create(Tx, &Admin{Username: "admin", Password: pwd.MustHash("123456")}); err != nil {
		return err
	}
	if err := orms.Create(Tx, &User{Username: "user", Password: pwd.MustHash("123456")}); err != nil {
		return err
	}
	return nil
}

func Migrate() error {
	return orms.AutoMigrate(Tx, Models...)
}

func Reset() error {
	return orms.Reset(conf.App.DBName, conf.App.DBDriver, conf.App.DBDsn)
}

func AutoCreate() error {
	return orms.AutoCreateE(conf.App.DBName, conf.App.DBDriver, conf.App.DBDsn)
}
