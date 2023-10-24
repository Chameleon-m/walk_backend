// Places API
//
// This is a places API.
//
//	Schemes: http
//	Host: localhost:8080
//	BasePath: /api/v1
//	Version: 0.0.1
//	Contact: Dmitry Korolev <korolev.d.l@yandex.ru> https://github.com/Chameleon-m
//
//	SecurityDefinitions:
//	    cookieAuth:
//	      type: apiKey
//	      in: cookie
//	      name: session_name
//
//	Consumes:
//	  - application/json
//
//	Produces:
//	  - application/json
//
// swagger:meta
package main

import (
	"flag"
	"log"
	"os"
	"walk_backend/config"
	"walk_backend/internal/pkg/app"
	"walk_backend/internal/pkg/util"
)

func main() {

	cfg, err := config.New()
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		log.Fatalf("Error: %s", err.Error())
	}

	if cfg.IsEnvDescription() {
		desc, err := cfg.GetEnvDescription()
		if err != nil {
			log.Fatalf("Config env description error: %s", err.Error())
		}
		log.Println(desc)
		os.Exit(0)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Config validating error: %s", err.Error())
	}

	if cfg.IsPrintConfig() {
		if err := util.PrintConfig(os.Stderr, cfg); err != nil {
			log.Fatalf("Config print error: %s", err.Error())
		}
		os.Exit(0)
	}

	if cfg.IsVerifyConfig() {
		log.Println("Config is valid")
		os.Exit(0)
	}

	server := app.New(&cfg.App)
	server.Run()
}
