// Package main is the entry point of webtplmst.
package main

import (
	"webtplmst/internal/conf"
	"webtplmst/internal/db"
	"webtplmst/internal/flag"
	"webtplmst/internal/srv"
	"webtplmst/internal/task"

	"github.com/natholdallas/natools4go/concur"
	"github.com/natholdallas/natools4go/vipers"
)

// @title						webtplmst
// @version					1.0
// @description				API Documentation
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
func main() {
	flag.PreRun()
	conf.Validate()
	db.Connect()
	flag.Run()
	vipers.Watch(conf.Reload, db.Reload)
	concur.Run(srv.Setup, task.Setup)
}
