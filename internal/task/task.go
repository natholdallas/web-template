// Package task to schedule task
package task

import (
	"webtplmst/internal/client"
	"webtplmst/internal/conf"
	"webtplmst/internal/db"

	"github.com/gofiber/fiber/v3/log"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

func Sync() {
	Rate()
}

func Setup() {
	schedule := cron.New(cron.WithSeconds())
	schedule.AddFunc("0 0 0,12 * * ?", Rate)
	schedule.Start()
}

func Rate() {
	if len(conf.App.RateCurrencies) == 0 {
		return
	}
	// Fetch every currency's rates first, failing fast before touching the DB
	// so a partial failure never leaves the table half-populated.
	all := make(map[string][]db.Rate, len(conf.App.RateCurrencies))
	for _, baseCode := range conf.App.RateCurrencies {
		log.Infof("caching rates %s ...", baseCode)
		rates, err := client.ExchangeRate(baseCode)
		if err != nil {
			log.Info("caching rates failed: ", err)
			return
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
	if err := db.Tx.Transaction(func(tx *gorm.DB) error {
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
	}); err != nil {
		log.Info("caching rates failed: ", err)
	}
}
