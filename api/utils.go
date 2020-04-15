package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/golang/gddo/httputil/header"
	uuid "github.com/satori/go.uuid"
)

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) (int, string) {
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			return http.StatusUnsupportedMediaType, msg
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return http.StatusBadRequest, msg

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			return http.StatusBadRequest, msg

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return http.StatusBadRequest, msg

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return http.StatusBadRequest, msg

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return http.StatusBadRequest, msg

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return http.StatusRequestEntityTooLarge, msg

		default:
			return http.StatusBadRequest, err.Error()
		}
	}

	if dec.More() {
		msg := "Request body must only contain a single JSON object"
		return http.StatusBadRequest, msg
	}

	return http.StatusOK, ""
}

func newCookie() (string, int) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return "", http.StatusInternalServerError
	}
	return uuid.String(), http.StatusOK
}
