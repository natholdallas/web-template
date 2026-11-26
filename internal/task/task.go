// Package task to schedule task
package task

import (
	"fmt"

	"github.com/gofiber/fiber/v3/log"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	"webgx/internal/client"
	"webgx/internal/conf"
	"webgx/internal/db"
)

// Sync runs the scheduled jobs once. It is the --sync CLI script.
func Sync() error {
	return Rate()
}

// Start schedules the recurring jobs and returns the running schedule so the
// caller can stop it on shutdown.
func Start() *cron.Cron {
	schedule := cron.New(cron.WithSeconds())
	if _, err := schedule.AddFunc("0 0 0,12 * * ?", func() {
		if err := Rate(); err != nil {
			log.Info("caching rates failed: ", err)
		}
	}); err != nil {
		panic(fmt.Errorf("schedule rate task: %w", err))
	}
	schedule.Start()
	return schedule
}

// Rate rebuilds the rates table from the configured exchange-rate provider. An
// unconfigured provider is a no-op, not an error.
func Rate() error {
	if conf.App.RateSite == "" || len(conf.App.RateCurrencies) == 0 {
		return nil
	}
	// Fetch every currency's rates first, failing fast before touching the DB
	// so a partial failure never leaves the table half-populated.
	all := make(map[string][]db.Rate, len(conf.App.RateCurrencies))
	for _, baseCode := range conf.App.RateCurrencies {
		log.Infof("caching rates %s ...", baseCode)
		rates, err := client.ExchangeRate(baseCode)
		if err != nil {
			return err
		}
		v := make([]db.Rate, 0, len(rates.Rates))
		for code, rate := range rates.Rates {
			v = append(v, db.Rate{BaseCode: rates.Code, Code: code, Rate: rate})
		}
		all[baseCode] = v
	}
	// Rebuild the table atomically. Use DELETE (not TRUNCATE): on MySQL
	// TRUNCATE is DDL and causes an implicit commit, which would defeat the
	// transaction and leave the table empty on a later failure.
	return db.Tx.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM rates").Error; err != nil {
			return err
		}
		for _, v := range all {
			if len(v) == 0 {
				continue
			}
			if err := tx.Create(&v).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
