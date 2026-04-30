// Package conf used to configuration vars on runtime
package conf

import (
	"github.com/fsnotify/fsnotify"
)

var (
	Flag *FlagConf = new(FlagConf)
	App  *AppConf  = new(AppConf)
)

func init() {
	LoadFlag()
	LoadApp()
}

func Reload(e fsnotify.Event) {
	LoadApp()
}
