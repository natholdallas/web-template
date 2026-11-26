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
	Username string `json:"username" validate:"required,max=50"`
	Password string `json:"password" validate:"required,min=4,max=72"`
} //	@name	AdmUserIn

type UserUpdateIn struct {
	Username string `json:"username" validate:"required,max=50"`
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
	id, err := internal.ParseID(c)
	if err != nil {
		return err
	}
	v, err := orms.First[db.User](db.Tx, id)
	if err != nil {
		return &fext.Fail{Code: internal.RecordNotFound, System: err, Status: fiber.StatusNotFound}
	}
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
	if orms.Exists[db.User](db.Tx.Where("username = ?", d.Username)) {
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
	id, err := internal.ParseID(c)
	if err != nil {
		return err
	}
	d, err := fext.BodyVarser[UserUpdateIn](c)
	if err != nil {
		return &fext.Fail{Code: internal.InvalidData, Message: err.Error()}
	}
	if orms.Exists[db.User](db.Tx.Where("username = ? AND id <> ?", d.Username, id)) {
		return &fext.Fail{Code: internal.UpdateFailed, Status: fiber.StatusConflict, Message: "username already exists"}
	}
	if err := orms.UpdatesByID[db.User](db.Tx, id, map[string]any{"username": d.Username}); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, System: err}
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
	id, err := internal.ParseID(c)
	if err != nil {
		return err
	}
	v, err := orms.First[db.User](db.Tx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &fext.Fail{Code: internal.RecordNotFound, Status: fiber.StatusNotFound}
		}
		return &fext.Fail{Code: internal.RemoveFailed, System: err}
	}
	affected, err := orms.DeleteByID[db.User](db.Tx, id)
	if err != nil {
		return &fext.Fail{Code: internal.RemoveFailed, System: err}
	}
	if affected == 0 {
		return &fext.Fail{Code: internal.RecordNotFound, Status: fiber.StatusNotFound}
	}
	// Uploads outlive the user row otherwise: every file would stay on disk with
	// no row pointing at it. RemoveMediaByURL below covers the avatar case where
	// no owning row exists.
	if media, ferr := orms.Find[db.Media](db.Tx, "user_id = ?", id); ferr != nil {
		log.Warn("find user media: ", ferr)
	} else {
		for _, m := range media {
			if rerr := db.RemoveByID(db.Tx, m.ID); rerr != nil {
				log.Warn("remove user media: ", rerr)
			}
		}
	}
	internal.RemoveMediaByURL(id, v.Avatar)
	if err := auth.Usr.RevokeAll(strs.FormatUint(id)); err != nil {
		log.Warn("revoke refresh tokens: ", err)
	}
	return nil
}

type ResetUserPasswordOut struct {
	Password string `json:"password"`
} //	@name	AdmResetUserPassword

// ResetUserPassword godoc
//
//	@Summary		Reset User password
//	@Description	Reset a user password to a new one, returned once. An empty password generates a random one.
//	@Tags			admUser
//	@ID				admResetUserPassword
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id		path		int					true	"User ID"
//	@Param			body	body		TargetPasswordIn	true	"Password object (empty password = random)"
//	@Success		200		{object}	ResetUserPasswordOut
//	@Failure		400		{object}	Fail
//	@Router			/adm/api/v1/user/{id}/reset-password [post]
func ResetUserPassword(c fiber.Ctx) error {
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
	if err := orms.UpdatesByID[db.User](db.Tx, id, map[string]any{"password": hash}); err != nil {
		return &fext.Fail{Code: internal.UpdateFailed, System: err}
	}
	if err := auth.Usr.RevokeAll(strs.FormatUint(id)); err != nil {
		log.Warn("revoke refresh tokens: ", err)
	}
	return c.JSON(ResetUserPasswordOut{Password: plain})
}
