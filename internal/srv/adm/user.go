package adm

import (
	"webtplmst/internal/db"
	"webtplmst/internal/srv/internal"

	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/orms"
	"github.com/natholdallas/natools4go/pwd"
	"github.com/natholdallas/natools4go/rands"
)

type UserQueries struct {
	orms.Sorter
	orms.Pagination
} //	@name	AdmUserQueries

type UsersPage struct {
	Total   int64     `json:"total"`
	Page    int64     `json:"page"`
	Content []db.User `json:"content"`
} //	@name	AdmUsersPage

type UserIn struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=4,max=20"`
} //	@name	AdmUserIn

type UserUpdateIn struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password"`
} //	@name	AdmUserUpdateIn

func (s *UserIn) Get() (*db.User, error) {
	hash, err := pwd.Hash(s.Password)
	if err != nil {
		return nil, err
	}
	return &db.User{
		Username: s.Username,
		Password: hash,
	}, nil
}

// ListUser godoc
//
//	@Summary	List Users
//	@Tags		admUser
//	@ID			admListUsers
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		query	query		UserQueries	false	"Query params"
//	@Success	200		{object}	UsersPage
//	@Router		/adm/api/v1/user [get]
func ListUser(c fiber.Ctx) error {
	q, err := fext.QueryParser[UserQueries](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	v := orms.QE[db.User](db.Tx).
		Scopes(q.Sorter.Scope).
		IPaginate(q.Pagination)
	return c.JSON(v)
}

// FindUser godoc
//
//	@Summary	Find User by ID
//	@Tags		admUser
//	@ID			admFindUser
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		id	path		int	true	"User ID"
//	@Success	200	{object}	User
//	@Router		/adm/api/v1/user/{id} [get]
func FindUser(c fiber.Ctx) error {
	v := orms.IFirst[db.User](db.Tx, c.Params("id"))
	return c.JSON(v)
}

// CreateUser godoc
//
//	@Summary	Create User
//	@Tags		admUser
//	@ID			admCreateUser
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		body	body		UserIn	true	"User object"
//	@Success	200		{object}	db.User
//	@Failure	400		{object}	Fail
//	@Router		/adm/api/v1/user [post]
func CreateUser(c fiber.Ctx) error {
	d, err := fext.BodyVarser[UserIn](c)
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

// UpdateUser godoc
//
//	@Summary	Update User
//	@Tags		admUser
//	@ID			admUpdateUser
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		id		path	int				true	"User ID"
//	@Param		body	body	UserUpdateIn	true	"User object"
//	@Success	200
//	@Failure	400	{object}	Fail
//	@Router		/adm/api/v1/user/{id} [put]
func UpdateUser(c fiber.Ctx) error {
	d, err := fext.BodyVarser[UserUpdateIn](c)
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
	if err := orms.UpdatesByID[db.User](db.Tx, c.Params("id"), values); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, Message: err.Error()}
	}
	return nil
}

// RemoveUser godoc
//
//	@Summary	Remove User
//	@Tags		admUser
//	@ID			admDeleteUser
//	@Accept		json
//	@Produce	json
//	@Security	ApiKeyAuth
//	@Param		id	path	int	true	"User ID"
//	@Success	200
//	@Router		/adm/api/v1/user/{id} [delete]
func RemoveUser(c fiber.Ctx) error {
	if err := orms.Delete[db.User](db.Tx, c.Params("id")); err != nil {
		return &fext.Fail{Code: internal.RemoveFailed, Message: err.Error()}
	}
	return nil
}

type ResetUserPasswordOut struct {
	Password string `json:"password"`
} //	@name	AdmResetUserPassword

// ResetUserPassword godoc
//
//	@Summary		Reset User password
//	@Description	Reset a user password to a new random one, returned once
//	@Tags			admUser
//	@ID				admResetUserPassword
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	ResetUserPasswordOut
//	@Failure		400	{object}	Fail
//	@Router			/adm/api/v1/user/{id}/reset-password [post]
func ResetUserPassword(c fiber.Ctx) error {
	plain := rands.Char(20)
	hash, err := pwd.Hash(plain)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	if err := orms.UpdatesByID[db.User](db.Tx, c.Params("id"), map[string]any{"password": hash}); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, Message: err.Error()}
	}
	return c.JSON(ResetUserPasswordOut{Password: plain})
}
