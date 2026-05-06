package admin

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
} //	@name	AdminQueries

// ListAdmin godoc
//
//	@Summary	List Admins
//	@Tags		Admin
//	@ID			list_admin
//	@Accept		json
//	@Produce	json
//	@Param		query	query	AdminQueries	false	"Query params"
//	@Success	200		{array}	db.Admin
//	@Router		/admin/api/v1/admin [get]
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
//	@Tags		Admin
//	@ID			find_admin
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int	true	"Admin ID"
//	@Success	200	{object}	db.Admin
//	@Router		/admin/api/v1/admin/{id} [get]
func FindAdmin(c fiber.Ctx) error {
	v := orms.IFirst[db.Admin](db.Tx)
	return c.JSON(v)
}

type AdminIn struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=4,max=20"`
} //	@name	AdminIn

func (s *AdminIn) Get() *db.Admin {
	return &db.Admin{
		Username: s.Username,
		Password: s.Password,
	}
}

// CreateAdmin godoc
//
//	@Summary	Create Admin
//	@Tags		Admin
//	@ID			create_admin
//	@Accept		json
//	@Produce	json
//	@Param		body	body		AdminIn	true	"Admin object"
//	@Success	200		{object}	db.Admin
//	@Router		/admin/api/admin/v1 [post]
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
//	@Tags		Admin
//	@ID			update_admin
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int		true	"Admin ID"
//	@Param		body	body		AdminIn	true	"Admin object"
//	@Success	200		{object}	string	"OK"
//	@Router		/admin/api/v1/admin/{id} [put]
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
//	@Tags		Admin
//	@ID			remove_admin
//	@Accept		json
//	@Produce	json
//	@Param		id	path	int	true	"Admin ID"
//	@Router		/admin/api/v1/admin/{id} [delete]
func RemoveAdmin(c fiber.Ctx) error {
	orms.Delete[db.Admin](db.Tx, c.Params("id"))
	return nil
}
