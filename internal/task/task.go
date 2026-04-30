// Package task to schedule task
package task

import (
	"io"
	"os"

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
	schedule.AddFunc("0 0 0 * * ?", Log)
	schedule.AddFunc("0 0 0,12 * * ?", Rate)
	schedule.Start()
}

func Log() {
	log.Info("setting up logger output...")
	conf.App.MkdirAll()
	stdout, err := os.OpenFile(conf.App.LogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o777)
	if err != nil {
		log.Error("set logger output failed: ", err)
		return
	}
	log.SetOutput(io.MultiWriter(os.Stdout, stdout))
}

func Rate() {
	db.Tx.Exec("truncate table rates")
	rate("CNY")
	rate("USD")
}

func rate(baseCode string) {
	log.Infof("caching rates %s ...", baseCode)
	rates, err := client.ExchangeRate(baseCode)
	if err != nil {
		log.Info("caching rates failed: ", err)
		return
	}
	r := []db.Rate{}
	for k, v := range rates.Rates {
		r = append(r, db.Rate{BaseCode: rates.Code, Code: k, Rate: v})
	}
	db.Tx.Create(r)
}
