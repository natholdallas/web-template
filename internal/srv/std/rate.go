package std

import (
	"webtplmst/internal/db"

	"github.com/gofiber/fiber/v3"
)

// FindRate godoc
//
//	@Summary		Get exchange rates
//	@Description	Get exchange rates by currency code
//	@Tags			Rate
//	@ID				find_rate
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]float64
//	@Router			/api/v1/rate/{code} [get]
func FindRate(c fiber.Ctx) error {
	v := map[string]float64{}
	rates := db.ListRateByBaseCode(db.Tx, c.Params("code"))
	for _, i := range rates {
		v[i.Code] = i.Rate
	}
	return c.JSON(v)
}
