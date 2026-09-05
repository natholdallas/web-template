package adm

import (
	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/pwd"
	"github.com/natholdallas/natools4go/rands"

	"webtplmst/internal/db"
	"webtplmst/internal/srv/internal"
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
	v := orms.IFirst[db.Admin](db.Tx, c.Params("id"))
	return c.JSON(v)
}

type AdminIn struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=4,max=20"`
} //	@name	AdmAdminIn

type AdminUpdateIn struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password"`
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
	v, err := d.Get()
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	if err := orms.Create(db.Tx, v); err != nil {
		return &fext.Fail{Code: internal.CreateFailed, Message: err.Error()}
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
	d, err := fext.BodyVarser[AdminUpdateIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	values := map[string]any{"username": d.Username}
	if d.Password != "" {
		hash, err := pwd.Hash(d.Password)
		if err != nil {
			return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
		}
		values["password"] = hash
	}
	if err := orms.UpdatesByID[db.Admin](db.Tx, c.Params("id"), values); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, Message: err.Error()}
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
	if err := orms.Delete[db.Admin](db.Tx, c.Params("id")); err != nil {
		return &fext.Fail{Code: internal.RemoveFailed, Message: err.Error()}
	}
	return nil
}

type ResetAdminPasswordOut struct {
	Password string `json:"password"`
} //	@name	AdmResetAdminPassword

// ResetAdminPassword godoc
//
//	@Summary		Reset Admin password
//	@Description	Reset an admin password to a new random one, returned once
//	@Tags			admAdmin
//	@ID				admResetAdminPassword
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		int	true	"Admin ID"
//	@Success		200	{object}	ResetAdminPasswordOut
//	@Failure		400	{object}	Fail
//	@Router			/adm/api/v1/admin/{id}/reset-password [post]
func ResetAdminPassword(c fiber.Ctx) error {
	plain := rands.Char(20)
	hash, err := pwd.Hash(plain)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	if err := orms.UpdatesByID[db.Admin](db.Tx, c.Params("id"), map[string]any{"password": hash}); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, Message: err.Error()}
	}
	return c.JSON(ResetAdminPasswordOut{Password: plain})
}
