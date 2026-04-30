package main

import (
	"webtplmst/internal/conf"
	"webtplmst/internal/flag"
	"webtplmst/internal/srv"
	"webtplmst/internal/task"

	"github.com/natholdallas/natools4go/concur"
	"github.com/natholdallas/natools4go/vipers"
)

// @title			Webtplmst
// @version		1.0
// @description	API Documentation
func main() {
	flag.Setup()
	vipers.NewUpdateEvent(srv.Reload, conf.Reload)
	vipers.Watch()
	concur.Run(srv.Setup, task.Setup)
}
