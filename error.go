package mvc

// HTTPError http error
type HTTPError struct {
	ResponseCode int
	Err          error
}

// NewHTTPError returns new HTTPError,
// filled with response code and original error
func NewHTTPError(responseCode int, err error) *HTTPError {
	return &HTTPError{
		ResponseCode: responseCode,
		Err:          err,
	}
}
