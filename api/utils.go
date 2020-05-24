package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/golang/gddo/httputil/header"
)

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) string {
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			return "Content-Type header is not application/json"
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		var msg string
		switch {
		case errors.As(err, &syntaxError):
			msg = fmt.Sprintf("request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg = fmt.Sprintf("request body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			msg = fmt.Sprintf("request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg = fmt.Sprintf("request body contains unknown field %s", fieldName)
		case errors.Is(err, io.EOF):
			msg = "request body must not be empty"
		case err.Error() == "http: request body too large":
			msg = "request body must not be larger than 1MB"
		default:
			msg = err.Error()
		}
		return msg
	}

	if decoder.More() {
		return "request body must only contain a single JSON object"
	}

	return ""
}
