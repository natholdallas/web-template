// Package conf used to configuration vars on runtime
package conf

import (
	"github.com/fsnotify/fsnotify"
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
	vipers.Validate(Flag)
	vipers.Validate(App)
}

func Reload(fsnotify.Event) {
	LoadApp()
}
