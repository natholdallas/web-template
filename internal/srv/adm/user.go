package adm

import (
	"webtplmst/internal/db"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/orms"
)

type UserQueries struct {
	orms.Sorter
	orms.Pagination
} //	@name	UserQueries

type UserIn struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=4,max=20"`
} //	@name	UserIn

func (s *UserIn) Get() *db.User {
	return &db.User{
		Username: s.Username,
		Password: s.Password,
	}
}

// ListUser godoc
//
//	@Summary	List Users
//	@Tags		RUser
//	@ID			admin__list_user
//	@Accept		json
//	@Produce	json
//	@Param		query	query	UserQueries	false	"Query params"
//	@Success	200		{array}	User
//	@Router		/admin/api/v1/user [get]
func ListUser(c fiber.Ctx) error {
	q, _ := fext.QueryParser[UserQueries](c)
	v := orms.QE[db.User](db.Tx).
		Scopes(q.Sorter.Scope).
		IPaginate(q.Pagination)
	return c.JSON(v)
}

// FindUser godoc
//
//	@Summary	Find User by ID
//	@Tags		RUser
//	@ID			admin__find_user
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int	true	"User ID"
//	@Success	200	{object}	User
//	@Router		/admin/api/v1/user/{id} [get]
func FindUser(c fiber.Ctx) error {
	v := orms.IFirst[db.User](db.Tx, c.Params("id"))
	return c.JSON(v)
}

// CreateUser godoc
//
//	@Summary	Create User
//	@Tags		RUser
//	@ID			admin__create_user
//	@Accept		json
//	@Produce	json
//	@Param		body	body		UserIn	true	"User object"
//	@Success	200		{object}	db.User
//	@Router		/admin/api/v1/user [post]
func CreateUser(c fiber.Ctx) error {
	d, err := fext.BodyVarser[UserIn](c)
	if err != nil {
		return err
	}
	v := d.Get()
	orms.Create(db.Tx, v)
	return nil
}

// UpdateUser godoc
//
//	@Summary	Update User
//	@Tags		RUser
//	@ID			admin__update_user
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int		true	"User ID"
//	@Param		body	body		UserIn	true	"User object"
//	@Success	200		{object}	string	"OK"
//	@Router		/admin/api/v1/user/{id} [put]
func UpdateUser(c fiber.Ctx) error {
	d, err := fext.BodyVarser[UserIn](c)
	if err != nil {
		return err
	}
	orms.UpdatesByID[db.User](db.Tx, c.Params("id"), map[string]any{
		"username": d.Username,
		"password": d.Password,
	})
	return nil
}

// RemoveUser godoc
//
//	@Summary	Remove User
//	@Tags		RUser
//	@ID			admin__remove_user
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int		true	"User ID"
//	@Success	200	{object}	string	"OK"
//	@Router		/admin/api/v1/user/{id} [delete]
func RemoveUser(c fiber.Ctx) error {
	orms.Delete[db.User](db.Tx, c.Params("id"))
	return nil
}
