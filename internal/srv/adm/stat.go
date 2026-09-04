package adm

import (
	"github.com/gofiber/fiber/v3"
	"webtplmst/internal/db"
)

type Stats struct {
	Admins int64 `json:"admins"`
	Users  int64 `json:"users"`
} //	@name	AdmStats

// GetStats godoc
//
//	@Summary	Get system stats
//	@Tags		admAdmin
//	@ID			admGetStats
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Success	200	{object}	Stats
//	@Router		/adm/api/v1/stat [get]
func GetStats(c fiber.Ctx) error {
	var s Stats
	db.Tx.Model(&db.Admin{}).Count(&s.Admins)
	db.Tx.Model(&db.User{}).Count(&s.Users)
	return c.JSON(s)
}
