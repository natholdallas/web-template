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

// @title			webtplmst
// @version		1.0
// @description	API Documentation
func main() {
	flag.Setup()
	vipers.Watch(conf.Reload, db.Reload)
	concur.Run(srv.Setup, task.Setup)
}
