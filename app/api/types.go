package api

import (
	"fmt"
	"net/http"
)

type middleware func(*Context, http.Handler) http.Handler

type httpResponse struct {
	Status    int         `json:"status"`
	ErrorName *string     `json:"error"`
	Message   *string     `json:"message"`
	Data      interface{} `json:"data"`
}

func (r httpResponse) Error() string {
	if r.ErrorName == nil || r.Message == nil {
		return ""
	}

	return fmt.Sprintf("%s (%d): %s", *r.ErrorName, r.Status, *r.Message)
}

func newHttpError(status int, name, msg string) httpResponse {
	res := httpResponse{
		Status:    status,
		ErrorName: &name,
		Message:   &msg,
	}

	return res
}

var (
	errInvalidParameters   = newHttpError(400, "InvalidParameters", "Invalid parameters")
	errRouteNotFound       = newHttpError(404, "RouteNotFound", "Route not found")
	errPeriodNotFound      = newHttpError(404, "PeriodNotFound", "Period not found")
	errMethodNotAllowed    = newHttpError(405, "MethodNotAllowed", "Method not allowed")
	errInternalServerError = newHttpError(500, "InternalServerError", "Internal server error")
)

type paginatedResponse struct {
	Page    uint64        `json:"page"`
	Size    uint64        `json:"size"`
	Results int           `json:"results"`
	Next    bool          `json:"next"`
	Items   []interface{} `json:"items"`
}
