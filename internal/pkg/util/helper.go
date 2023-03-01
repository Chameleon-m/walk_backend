package util

import (
	"net/http"
)

func MakeURL(request *http.Request, uri string) string {

	url := request.URL.Scheme + "://"
	if request.URL.User.String() != "" {
		url += request.URL.User.String() + "@"
	}
	return url + request.URL.Host + uri
}
