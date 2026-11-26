// Package main is the entry point of webgx.
package main

import "webgx/internal/app"

// @title						webgx
// @version					1.0
// @description				API Documentation
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
func main() {
	app.Run()
}
