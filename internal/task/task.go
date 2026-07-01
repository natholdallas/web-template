// Package task to schedule task
package task

import (
	"webtplmst/internal/client"
	"webtplmst/internal/conf"
	"webtplmst/internal/db"

	"github.com/gofiber/fiber/v3/log"
	"github.com/robfig/cron/v3"
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
	db.Tx.Exec("TRUNCATE TABLE rates")
	for _, baseCode := range conf.App.RateCurrencies {
		log.Infof("caching rates %s ...", baseCode)
		rates, err := client.ExchangeRate(baseCode)
		if err != nil {
			log.Info("caching rates failed: ", err)
			return
		}
		v := []db.Rate{}
		for code, rate := range rates.Rates {
			v = append(v, db.Rate{BaseCode: rates.Code, Code: code, Rate: rate})
		}
		db.Tx.Create(v)
	}
}
