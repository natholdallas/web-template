package adm

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/pwd"
	"github.com/natholdallas/natools4go/rands"
	"github.com/natholdallas/natools4go/strs"
	"gorm.io/gorm"

	"webgx/internal/auth"
	"webgx/internal/db"
	"webgx/internal/srv/internal"
)

type AdminQueries struct {
	orms.Sorter
	orms.Pagination
} //	@name	AdmAdminQueries

type AdminsPage struct {
	Total   int64      `json:"total"`
	Page    int64      `json:"page"`
	Content []db.Admin `json:"content"`
} //	@name	AdmAdminsPage

// ListAdmin godoc
//
//	@Summary	List Admins
//	@Tags		admAdmin
//	@ID			admListAdmins
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		query	query		AdminQueries	false	"Query params"
//	@Success	200		{object}	AdminsPage
//	@Router		/adm/api/v1/admin [get]
func ListAdmin(c fiber.Ctx) error {
	q, err := fext.QueryParser[AdminQueries](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	v := orms.Q[db.Admin](db.Tx).
		Scopes(q.Sorter.Scope).
		IPaginate(q.Pagination)
	return c.JSON(v)
}

// FindAdmin godoc
//
//	@Summary	Find Admin by ID
//	@Tags		admAdmin
//	@ID			admFindAdmin
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		id	path		int	true	"Admin ID"
//	@Success	200	{object}	db.Admin
//	@Router		/adm/api/v1/admin/{id} [get]
func FindAdmin(c fiber.Ctx) error {
	id, err := internal.ParseID(c)
	if err != nil {
		return err
	}
	v, err := orms.First[db.Admin](db.Tx, id)
	if err != nil {
		return &fext.Fail{Code: internal.RecordNotFound, System: err, Status: fiber.StatusNotFound}
	}
	return c.JSON(v)
}

type AdminIn struct {
	Username string `json:"username" validate:"required,max=50"`
	Password string `json:"password" validate:"required,min=4,max=72"`
} //	@name	AdmAdminIn

type AdminUpdateIn struct {
	Username string `json:"username" validate:"required,max=50"`
} //	@name	AdmAdminUpdateIn

func (s *AdminIn) Get() (*db.Admin, error) {
	hash, err := pwd.Hash(s.Password)
	if err != nil {
		return nil, err
	}
	return &db.Admin{
		Username: s.Username,
		Password: hash,
	}, nil
}

// CreateAdmin godoc
//
//	@Summary	Create Admin
//	@Tags		admAdmin
//	@ID			admCreateAdmin
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		body	body		AdminIn	true	"Admin object"
//	@Success	200		{object}	db.Admin
//	@Failure	400		{object}	Fail
//	@Router		/adm/api/v1/admin [post]
func CreateAdmin(c fiber.Ctx) error {
	d, err := fext.BodyVarser[AdminIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	if orms.Exists[db.Admin](db.Tx.Where("username = ?", d.Username)) {
		return &fext.Fail{Code: internal.CreateFailed, Status: fiber.StatusConflict, Message: "username already exists"}
	}
	v, err := d.Get()
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	if err := orms.Create(db.Tx, v); err != nil {
		return &fext.Fail{Code: internal.CreateFailed, System: err}
	}
	return nil
}

// UpdateAdmin godoc
//
//	@Summary	Update Admin
//	@Tags		admAdmin
//	@ID			admUpdateAdmin
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		id		path	int				true	"Admin ID"
//	@Param		body	body	AdminUpdateIn	true	"Admin object"
//	@Success	200
//	@Failure	400	{object}	Fail
//	@Router		/adm/api/v1/admin/{id} [put]
func UpdateAdmin(c fiber.Ctx) error {
	id, err := internal.ParseID(c)
	if err != nil {
		return err
	}
	d, err := fext.BodyVarser[AdminUpdateIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	if orms.Exists[db.Admin](db.Tx.Where("username = ? AND id <> ?", d.Username, id)) {
		return &fext.Fail{Code: internal.UpdateFailed, Status: fiber.StatusConflict, Message: "username already exists"}
	}
	if err := orms.UpdatesByID[db.Admin](db.Tx, id, map[string]any{"username": d.Username}); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, System: err}
	}
	return nil
}

// RemoveAdmin godoc
//
//	@Summary	Remove Admin
//	@Tags		admAdmin
//	@ID			admDeleteAdmin
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		id	path	int	true	"Admin ID"
//	@Success	200
//	@Router		/adm/api/v1/admin/{id} [delete]
func RemoveAdmin(c fiber.Ctx) error {
	id, err := internal.ParseID(c)
	if err != nil {
		return err
	}
	v, err := orms.First[db.Admin](db.Tx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &fext.Fail{Code: internal.RecordNotFound, Status: fiber.StatusNotFound}
		}
		return &fext.Fail{Code: internal.RemoveFailed, System: err}
	}
	affected, err := orms.DeleteByID[db.Admin](db.Tx, id)
	if err != nil {
		return &fext.Fail{Code: internal.RemoveFailed, System: err}
	}
	if affected == 0 {
		return &fext.Fail{Code: internal.RecordNotFound, Status: fiber.StatusNotFound}
	}
	internal.RemoveMediaByURL(id, v.Avatar)
	if err := auth.Adm.RevokeAll(strs.FormatUint(id)); err != nil {
		log.Warn("revoke refresh tokens: ", err)
	}
	return nil
}

// TargetPasswordIn carries an optional replacement password. An empty value
// means "generate a random password".
type TargetPasswordIn struct {
	Password string `json:"password" validate:"omitempty,min=4,max=72"`
} //	@name	AdmTargetPasswordIn

type ResetAdminPasswordOut struct {
	Password string `json:"password"`
} //	@name	AdmResetAdminPassword

// ResetAdminPassword godoc
//
//	@Summary		Reset Admin password
//	@Description	Reset an admin password to a new one, returned once. An empty password generates a random one.
//	@Tags			admAdmin
//	@ID				admResetAdminPassword
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id		path		int					true	"Admin ID"
//	@Param			body	body		TargetPasswordIn	true	"Password object (empty password = random)"
//	@Success		200		{object}	ResetAdminPasswordOut
//	@Failure		400		{object}	Fail
//	@Router			/adm/api/v1/admin/{id}/reset-password [post]
func ResetAdminPassword(c fiber.Ctx) error {
	id, err := internal.ParseID(c)
	if err != nil {
		return err
	}
	d, err := fext.BodyVarser[TargetPasswordIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	plain := d.Password
	if plain == "" {
		plain = rands.Char(20)
	}
	hash, err := pwd.Hash(plain)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	if err := orms.UpdatesByID[db.Admin](db.Tx, id, map[string]any{"password": hash}); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, System: err}
	}
	if err := auth.Adm.RevokeAll(strs.FormatUint(id)); err != nil {
		log.Warn("revoke refresh tokens: ", err)
	}
	return c.JSON(ResetAdminPasswordOut{Password: plain})
}
