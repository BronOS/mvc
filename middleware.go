package mvc

import (
	"net/http"
)

// MiddlewareInterface server middleware interface.
// In case when error returned, calls http.Error with response code and error message
type MiddlewareInterface interface {
	Handle(w http.ResponseWriter, r *http.Request) *HTTPError
}
