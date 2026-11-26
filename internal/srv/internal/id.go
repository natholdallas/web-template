package internal

import (
	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/strs"
)

// ParseID parses a numeric route parameter into uint.
//
// Handlers must never hand c.Params("id") straight to orms.First/Delete: GORM
// treats any string containing a space or "?" as a raw WHERE clause, so
// "/user/1%20or%201%3D1" would delete or reveal every row. Parsing first means a
// non-numeric id fails as a 400 instead of reaching the query builder.
//
// Usage:
//
//	id, err := internal.ParseID(c)
//	if err != nil {
//		return err
//	}
func ParseID(c fiber.Ctx, name ...string) (uint, error) {
	param := "id"
	if len(name) > 0 && name[0] != "" {
		param = name[0]
	}
	raw := c.Params(param)
	v, err := strs.ParseUintE[uint](raw)
	if err != nil {
		return 0, &fext.Fail{Code: InvalidData, Message: "invalid " + param}
	}
	if v == 0 {
		// IDs are auto-increment starting at 1; 0 is never a real record and
		// would make GORM's "id = 0" condition ambiguous with "no condition".
		return 0, &fext.Fail{Code: InvalidData, Message: "invalid " + param}
	}
	return v, nil
}
