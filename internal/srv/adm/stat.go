package adm

import (
	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"

	"webgx/internal/db"
	"webgx/internal/srv/internal"
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
	if err := db.Tx.Model(&db.Admin{}).Count(&s.Admins).Error; err != nil {
		return &fext.Fail{Code: internal.OperationFailed, System: err}
	}
	if err := db.Tx.Model(&db.User{}).Count(&s.Users).Error; err != nil {
		return &fext.Fail{Code: internal.OperationFailed, System: err}
	}
	return c.JSON(s)
}
