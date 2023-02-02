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
	"os"
	"walk_backend/internal/pkg/app"
	httpserver "walk_backend/internal/pkg/component/http_server"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true})

	var err error
	var server httpserver.ServerInterface

	server, err = app.New()
	if err != nil {
		log.Fatal().Err(err).Caller(1).Send()
	}

	err = server.Run()
	if err != nil {
		log.Fatal().Err(err).Caller(1).Send()
	}
}
