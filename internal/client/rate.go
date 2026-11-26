package client

import (
	"errors"

	"webgx/internal/conf"
)

type Rate struct {
	Result string             `json:"result"`
	Code   string             `json:"base_code"`
	Rates  map[string]float64 `json:"conversion_rates"`
}

func ExchangeRate(currency string) (*Rate, error) {
	data := Rate{Rates: make(map[string]float64)}
	_, err := requestClient().R().
		SetResult(&data).
		Get(conf.App.RateSite + "/" + currency)
	if err != nil {
		return nil, err
	}
	if data.Result != "success" {
		return nil, errors.New("get exchange rate error")
	}
	return &data, nil
}
