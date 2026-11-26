package adm

import (
	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/orms"

	"webgx/internal/conf"
	"webgx/internal/db"
	"webgx/internal/srv/internal"
	"webgx/internal/task"
)

type RateQueries struct {
	orms.Sorter
	orms.Pagination
} //	@name	AdmRateQueries

// RatesPage adds the provider state to the paginated rates so the console can
// disable its sync action without a second request.
type RatesPage struct {
	Enabled    bool      `json:"enabled"`
	Currencies []string  `json:"currencies"`
	Total      int64     `json:"total"`
	Page       int64     `json:"page"`
	Content    []db.Rate `json:"content"`
} //	@name	AdmRatesPage

type SyncRateOut struct {
	Synced bool `json:"synced"`
} //	@name	AdmSyncRate

// rateEnabled reports whether an exchange-rate provider is configured. The
// console disables its sync button when this is false.
func rateEnabled() bool {
	return conf.App.RateSite != "" && len(conf.App.RateCurrencies) > 0
}

// ListRate godoc
//
//	@Summary	List cached exchange rates
//	@Tags		admRate
//	@ID			admListRates
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		query	query		RateQueries	false	"Query params"
//	@Success	200		{object}	RatesPage
//	@Failure	400		{object}	Fail
//	@Router		/adm/api/v1/rate [get]
func ListRate(c fiber.Ctx) error {
	q, err := fext.QueryParser[RateQueries](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	v := orms.QE[db.Rate](db.Tx).
		Scopes(q.Sorter.Scope).
		IPaginate(q.Pagination)
	currencies := conf.App.RateCurrencies
	if currencies == nil {
		currencies = []string{}
	}
	return c.JSON(RatesPage{
		Enabled:    rateEnabled(),
		Currencies: currencies,
		Total:      v.Total,
		Page:       v.Page,
		Content:    v.Content,
	})
}

// SyncRate godoc
//
//	@Summary		Rebuild the cached exchange rates
//	@Description	Fetches every configured base currency and replaces the cached rate rows
//	@Tags			admRate
//	@ID				admSyncRate
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	SyncRateOut
//	@Failure		400	{object}	Fail
//	@Router			/adm/api/v1/rate/sync [post]
func SyncRate(c fiber.Ctx) error {
	if !rateEnabled() {
		return &fext.Fail{Code: internal.InvalidData, Message: "exchange rate not configured"}
	}
	if err := task.Rate(); err != nil {
		return &fext.Fail{Code: internal.OperationFailed, System: err}
	}
	return c.JSON(SyncRateOut{Synced: true})
}
