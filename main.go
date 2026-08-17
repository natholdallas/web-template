// Package main is the entry point of webtplmst.
//
// Startup flow (package execution order):
//
//	[1] internal/conf init()
//	     LoadFlag()  — parse all CLI flags (--remake-secret, --adm, --usr, --rstdb, --migration, --sync, --mock)
//	     LoadApp()   — load conf.toml configuration into conf.App
//
//	[2] main()
//	     flag.PreRun()   — execute pre-db scripts (only --remake-secret)
//	                       exits via os.Exit(0) without triggering database connection
//	     db.Connect()    — connect to database + Redis; runs auto-migration if DBAutoMigrate is set
//	     conf.Validate() — validate config integrity (secrets must be 32-char, etc.)
//	     flag.Run()      — execute db-dependent scripts (--rstdb/--migration/--adm/--usr/--sync/--mock)
//	                       each handler exits via os.Exit(0)
//	     vipers.Watch()  — watch conf.toml for changes, hot-reload config and log level
//	     concur.Run(srv.Setup, task.Setup)
//	                     — start Fiber HTTP server (port 8080) and cron scheduler (rate sync at 0:00 & 12:00) in parallel
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

// @title          webtplmst
// @version        1.0
// @description    API Documentation
func main() {
	flag.PreRun()
	db.Connect()
	conf.Validate()
	flag.Run()
	vipers.Watch(conf.Reload, db.Reload)
	concur.Run(srv.Setup, task.Setup)
}
