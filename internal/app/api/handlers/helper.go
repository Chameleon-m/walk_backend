package handlers

import (
	"net/http"
)

// makeURL
func makeURL(request *http.Request, uri string) string {

	url := request.URL.Scheme + "://"
	if request.URL.User.String() != "" {
		url += request.URL.User.String() + "@"
	}
	return url + request.URL.Host + uri
}
