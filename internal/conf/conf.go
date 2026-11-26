// Package conf used to configuration vars on runtime
package conf

import (
	"fmt"
	"strings"

	"github.com/natholdallas/natools4go/vipers"
)

var (
	Flag *FlagConf = new(FlagConf)
	App  *AppConf  = new(AppConf)
)

// Load parses CLI flags and reads conf.toml. It must run once from the process
// entrypoint before Validate or any package reads Flag/App.
func Load() {
	LoadFlag()
	LoadApp()
}

func Validate() {
	vipers.MustValidate(Flag)
	vipers.MustValidate(App)
	switch App.SMTPTLS {
	case SMTPTLSAuto, SMTPTLSStartTLS, SMTPTLSSSL, SMTPTLSNone:
	default:
		panic(fmt.Errorf("smtp.tls: unsupported value %q (want auto, starttls, ssl or none)", App.SMTPTLS))
	}
	validateDatabase()
	// The body limit caps the whole request, so it must fit the largest single
	// file we accept; otherwise the upload is rejected before its own check.
	if App.BodyLimit < max(App.MediaMaxSize, App.AvatarMaxSize) {
		panic(fmt.Errorf(
			"app.body-limit (%d MB) must be >= resources.media-max-size and resources.media-avatar-max-size",
			App.BodyLimit,
		))
	}
	// These are used as sizes, durations and rate-limit windows. Zero or a
	// negative value disables the guard silently (a zero TTL mints instantly
	// expired tokens, a zero size cap rejects every upload), so refuse to boot.
	positive := map[string]int{
		"app.body-limit":                  App.BodyLimit,
		"jwt.access-minutes":              App.AccessMinutes,
		"jwt.refresh-hours":               App.RefreshHours,
		"resources.media-max-size":        App.MediaMaxSize,
		"resources.media-avatar-max-size": App.AvatarMaxSize,
	}
	for key, value := range positive {
		if value <= 0 {
			panic(fmt.Errorf("%s must be > 0", key))
		}
	}
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
