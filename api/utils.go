package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Gimulator/Gimulator/types"
	"github.com/Gimulator/protobuf/go/api"
	"github.com/golang/gddo/httputil/header"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func validateKey(key *api.Key, method types.Method) error {
	if key == nil {
		return status.Errorf(codes.InvalidArgument, "the key cannot be nil/null")
	}

	switch method {
	case types.GetMethod:
		return validateGetKey(key)
	case types.GetAllMethod:
		return validateGetAllKey(key)
	case types.PutMethod:
		return validatePutKey(key)
	case types.DeleteMethod:
		return validateDeleteKey(key)
	case types.DeleteAllMethod:
		return validateDeleteAllKey(key)
	case types.WatchMethod:
		return validateWatchKey(key)
	default:
		//TODO
	}
	return nil
}

func validateGetKey(key *api.Key) error {
	if key.Type == "" {
		return status.Errorf(codes.InvalidArgument, "The type field of a key in the GET request cannot be empty")
	}
	if key.Name == "" {
		return status.Errorf(codes.InvalidArgument, "The name field of a key in the GET request cannot be empty")
	}
	if key.Namespace == "" {
		return status.Errorf(codes.InvalidArgument, "The namespace field of a key in the GET request cannot be empty")
	}
	return nil
}

func validateGetAllKey(key *api.Key) error {
	return nil
}

func validatePutKey(key *api.Key) error {
	if key.Type == "" {
		return status.Errorf(codes.InvalidArgument, "the Type field of a key in the PUT request cannot be empty")
	}
	if key.Name == "" {
		return status.Errorf(codes.InvalidArgument, "the Name field of a key in the PUT request cannot be empty")
	}
	if key.Namespace == "" {
		return status.Errorf(codes.InvalidArgument, "the Namespace field of a key in the PUT request cannot be empty")
	}
	return nil
}

func validateDeleteKey(key *api.Key) error {
	if key.Type == "" {
		return status.Errorf(codes.InvalidArgument, "the Type field of a key in the DELETE request cannot be empty")
	}
	if key.Name == "" {
		return status.Errorf(codes.InvalidArgument, "the Name field of a key in the DELETE request cannot be empty")
	}
	if key.Namespace == "" {
		return status.Errorf(codes.InvalidArgument, "the Namespace field of a key in the DELETE request cannot be empty")
	}
	return nil
}

func validateDeleteAllKey(key *api.Key) error {
	return nil
}

func validateWatchKey(key *api.Key) error {
	return nil
}
