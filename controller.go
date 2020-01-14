package mvc

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/astaxie/beego/validation"
	"github.com/gorilla/mux"
	modSchema "github.com/gorilla/schema"
)

// ControllerInterface calls "Action" method on received request.
// In case when HTTPError is returned, calls http.Error with HTTPError.ResponseCode and error message.
type ControllerInterface interface {
	Action(w http.ResponseWriter, r *http.Request) *HTTPError
}

// AbstractController abstract controller provides some of functionality related to request and response,
// such as ScanQuery, ScanVars, ScanForm, etc...
type AbstractController struct {
	schemaDecoder *modSchema.Decoder
}

func (c *AbstractController) getSchemaDecoder() *modSchema.Decoder {
	if c.schemaDecoder == nil {
		c.schemaDecoder = modSchema.NewDecoder()
	}

	return c.schemaDecoder
}

// ScanQuery scans query/URI string into schema struct based on gorilla/schema lib
func (c *AbstractController) ScanQuery(r *http.Request, schema interface{}) *HTTPError {
	if err := c.getSchemaDecoder().Decode(schema, r.URL.Query()); err != nil {
		return NewHTTPError(http.StatusBadRequest, err)
	}

	return nil
}

// ScanVars scans route's variables into schema struct based on MUX router lib
func (c *AbstractController) ScanVars(r *http.Request, schema interface{}) *HTTPError {
	vars := mux.Vars(r)
	mVars := make(map[string][]string)

	for k, v := range vars {
		mVars[k] = []string{v}
	}

	if err := c.getSchemaDecoder().Decode(schema, mVars); err != nil {
		return NewHTTPError(http.StatusBadRequest, err)
	}

	return nil
}

// ScanForm scans body form variables into schema struct.
func (c *AbstractController) ScanForm(r *http.Request, schema interface{}) *HTTPError {
	if err := r.ParseForm(); err != nil {
		return NewHTTPError(http.StatusBadRequest, err)
	}

	if err := c.getSchemaDecoder().Decode(schema, r.PostForm); err != nil {
		return NewHTTPError(http.StatusBadRequest, err)
	}

	return nil
}

// AbstractJSONController provides functionality related to JSON request and response,
// such as parsing/scanning of JSON body, writing schema struct into response body as a JSON string
type AbstractJSONController struct {
	AbstractController
}

// ScanJSONBody parse body as JSON string into schema struct.
func (c *AbstractJSONController) ScanJSONBody(r *http.Request, schema interface{}) *HTTPError {
	if err := json.NewDecoder(r.Body).Decode(schema); err != nil {
		return NewHTTPError(http.StatusBadRequest, err)
	}

	valid := validation.Validation{}
	b, err := valid.Valid(schema)

	if err != nil {
		return NewHTTPError(http.StatusBadRequest, err)
	}

	if !b {
		return NewHTTPError(http.StatusBadRequest, errors.New("Schema validation error"))
	}

	return nil
}

// WriteJSONResponse writes schema struct into response body as a JSON string
func (c *AbstractJSONController) WriteJSONResponse(w http.ResponseWriter, schema interface{}, responseCode *int) *HTTPError {
	b, err := json.Marshal(schema)
	if err != nil {
		return NewHTTPError(http.StatusInternalServerError, err)
	}

	w.Header().Set("Content-Type", "application/json")

	code := 200
	if responseCode != nil {
		code = *responseCode
	}

	w.WriteHeader(code)
	w.Write(b)

	return nil
}
