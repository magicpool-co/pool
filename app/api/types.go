package api

import (
	"fmt"
	"net/http"
)

var (
	errInvalidParameters   = newHttpError(400, "InvalidParameters", "Invalid parameters")
	errInvalidJSONBody     = newHttpError(400, "InvalidJSONBody", "Invalid JSON body")
	errBodyTooLarge        = newHttpError(400, "BodyTooLarge", "Body too large")
	errTooManyMiners       = newHttpError(400, "TooManyMiners", "Too many miners requested")
	errInvalidThreshold    = newHttpError(400, "InvalidThreshold", "Invalid payout threshold")
	errThresholdTooSmall   = newHttpError(400, "ThresholdTooSmall", "Threshold too small")
	errThresholdTooBig     = newHttpError(400, "ThresholdTooBig", "Threshold too big")
	errIncorrectIPAddress  = newHttpError(403, "IncorrectIPAddress", "Incorrect IP address")
	errRouteNotFound       = newHttpError(404, "RouteNotFound", "Route not found")
	errChainNotFound       = newHttpError(404, "ChainNotFound", "Chain not found")
	errPeriodNotFound      = newHttpError(404, "PeriodNotFound", "Period not found")
	errMetricNotFound      = newHttpError(404, "MetricNotFound", "Metric not found")
	errMinerNotFound       = newHttpError(404, "MinerNotFound", "Miner not found")
	errWorkerNotFound      = newHttpError(404, "WorkerNotFound", "Worker not found")
	errMethodNotAllowed    = newHttpError(405, "MethodNotAllowed", "Method not allowed")
	errInternalServerError = newHttpError(500, "InternalServerError", "Internal server error")
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

func (r httpResponse) Equals(r2 httpResponse) bool {
	if (r.ErrorName == nil) != (r2.ErrorName == nil) {
		return false
	} else if r.ErrorName == nil && r2.ErrorName == nil {
		return false
	}

	return *r.ErrorName == *r2.ErrorName
}

func newHttpError(status int, name, msg string) httpResponse {
	res := httpResponse{
		Status:    status,
		ErrorName: &name,
		Message:   &msg,
	}

	return res
}

type paginatedResponse struct {
	Page    uint64        `json:"page"`
	Size    uint64        `json:"size"`
	Results uint64        `json:"results"`
	Next    bool          `json:"next"`
	Items   []interface{} `json:"items"`
}
