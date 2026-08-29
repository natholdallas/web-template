package adm

import (
	"webtplmst/internal/db"
	"webtplmst/internal/srv/internal"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/orms"
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
	v := orms.IFirst[db.Admin](db.Tx)
	return c.JSON(v)
}

type AdminIn struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=4,max=20"`
} //	@name	AdmAdminIn

func (s *AdminIn) Get() *db.Admin {
	return &db.Admin{
		Username: s.Username,
		Password: s.Password,
	}
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
	v := d.Get()
	orms.Create(db.Tx, v)
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
//	@Param		id		path	int		true	"Admin ID"
//	@Param		body	body	AdminIn	true	"Admin object"
//	@Success	200
//	@Failure	400	{object}	Fail
//	@Router		/adm/api/v1/admin/{id} [put]
func UpdateAdmin(c fiber.Ctx) error {
	d, err := fext.BodyVarser[AdminIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	v := d.Get()
	orms.UpdatesByID[db.Admin](db.Tx, c.Params("id"), v)
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
	orms.Delete[db.Admin](db.Tx, c.Params("id"))
	return nil
}

type ProfileIn struct {
	Username string `json:"username" validate:"required"`
} //	@name	AdmProfileIn

// UpdateProfile godoc
//
//	@Summary	Update current admin profile
//	@Tags		admAdmin
//	@ID			admUpdateProfile
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		body	body	ProfileIn	true	"Profile object"
//	@Success	200
//	@Router		/adm/api/v1/admin/profile [put]
func UpdateProfile(c fiber.Ctx) error {
	d, err := fext.BodyVarser[ProfileIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	claims := jwt.Claims(c)
	orms.UpdatesByID[db.Admin](db.Tx, claims.ID, map[string]any{
		"username": d.Username,
	})
	return nil
}

type ResetPasswordIn struct {
	Old string `json:"old" validate:"required,min=4,max=50"`
	New string `json:"new" validate:"required,min=4,max=50"`
} //	@name	AdmResetPasswordIn

// ResetPassword godoc
//
//	@Summary	Reset current admin password
//	@Tags		admAdmin
//	@ID			admResetPassword
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		body	body	ResetPasswordIn	true	"Password object"
//	@Success	200
//	@Failure	400	{object}	Fail
//	@Router		/adm/api/v1/admin/password [put]
func ResetPassword(c fiber.Ctx) error {
	d, err := fext.BodyVarser[ResetPasswordIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	claims := jwt.Claims(c)
	result := db.Tx.
		Model(&db.Admin{}).
		Where("id = ? AND password = ?", claims.ID, d.Old).
		Update("password", d.New)
	if result.RowsAffected == 0 {
		return &fext.Fail{Code: internal.OperationFailed, Message: "incorrect old password"}
	}
	return nil
}
