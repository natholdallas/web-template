// Package app owns the backend process lifecycle: configuration,
// infrastructure, one-shot CLI scripts, background jobs and the HTTP server.
package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"webgx/internal/auth"
	"webgx/internal/conf"
	"webgx/internal/db"
	"webgx/internal/flag"
	"webgx/internal/srv"
	"webgx/internal/task"
)

// Run boots the process. One-shot flags (--gensec, --db-*, --adm, ...) exit
// inside the flag helpers; only a normal boot reaches the HTTP server.
func Run() {
	conf.Load()
	// --gensec only prints fresh secrets; validating and connecting afterwards
	// would be pointless and would fail on the still-empty configuration.
	flag.PreRun()
	conf.Validate()
	auth.Setup()
	db.Connect()
	defer db.Close()
	flag.Run()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	schedule := task.Start()
	defer func() { <-schedule.Stop().Done() }()

	if err := srv.Run(ctx); err != nil {
		panic(err)
	}
}
