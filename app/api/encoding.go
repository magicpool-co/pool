package api

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/goccy/go-json"
)

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if r.Header.Get("Content-Type") != "application/json" {
		return errInvalidJSONBody
	}

	// enforce 25kb max body size
	r.Body = http.MaxBytesReader(w, r.Body, 25600)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return errInvalidJSONBody
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errInvalidJSONBody
		case errors.As(err, &unmarshalTypeError):
			return errInvalidJSONBody
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			return errInvalidJSONBody
		case errors.Is(err, io.EOF):
			return errInvalidJSONBody
		case err.Error() == "http: request body too large":
			return errBodyTooLarge

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errInvalidJSONBody
	}

	return nil
}
