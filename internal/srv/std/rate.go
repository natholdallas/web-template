package std

import (
	"github.com/gofiber/fiber/v3"

	"webtplmst/internal/db"
)

// FindRate godoc
//
//	@Summary		Get exchange rates
//	@Description	Get exchange rates by currency code
//	@Tags			Rate
//	@ID				findRate
//	@Accept			json
//	@Produce		json
//	@Param			code	path		string	true	"Currency code"
//	@Success		200		{object}	map[string]float64
//	@Router			/api/v1/rate/{code} [get]
func FindRate(c fiber.Ctx) error {
	v := map[string]float64{}
	rates := db.ListRateByBaseCode(db.Tx, c.Params("code"))
	for _, i := range rates {
		v[i.Code] = i.Rate
	}
	return c.JSON(v)
}
