// Package conf used to configuration vars on runtime
package conf

import (
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/natholdallas/natools4go/narder"
	"github.com/natholdallas/natools4go/vipers"
)

var (
	Flag *FlagConf = new(FlagConf)
	App  *AppConf  = new(AppConf)
)

func init() {
	LoadFlag()
	LoadApp()
}

func Validate() {
	vipers.MustValidate(Flag)
	vipers.MustValidate(App)
	validateDatabase()
}

func Reload(fsnotify.Event) {
	LoadApp()
	// Do NOT dump App here: it embeds SecretAdm/SecretUsr and every SMTP/WeChat
	// credential, and this runs on every conf.toml save.
	narder.Info("configuration reloaded")
}

// validateDatabase enforces the connection fields each driver actually needs.
// A blanket `validate:"required"` on host/port/credentials made sqlite unusable:
// a file database needs none of them.
func validateDatabase() {
	switch App.DBDriver {
	case "sqlite", "sqlite3":
		return
	}
	var missing []string
	if App.DBHost == "" {
		missing = append(missing, "db.host")
	}
	if App.DBPort == "" {
		missing = append(missing, "db.port")
	}
	if len(missing) > 0 {
		panic(fmt.Errorf("missing required configuration:\n\t%s", strings.Join(missing, "\n\t")))
	}
}
