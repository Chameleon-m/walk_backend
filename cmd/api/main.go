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
	"walk_backend/internal/pkg/app"
)

func main() {
	server := app.New()
	server.Run()
}
