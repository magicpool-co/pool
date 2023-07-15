package api

import (
	"fmt"
	"net/http"
)

var (
	errInvalidParameters     = newHttpError(400, "InvalidParameters", "Invalid parameters", true)
	errInvalidJSONBody       = newHttpError(400, "InvalidJSONBody", "Invalid JSON body", true)
	errBodyTooLarge          = newHttpError(400, "BodyTooLarge", "Body too large", true)
	errStreamingNotSupported = newHttpError(400, "StreamingNotSupported", "Streaming not supported", true)
	errTooManyMiners         = newHttpError(400, "TooManyMiners", "Too many miners requested", true)
	errInvalidEmail          = newHttpError(400, "InvalidEmail", "Invalid email address", true)
	errInvalidThreshold      = newHttpError(400, "InvalidThreshold", "Invalid payout threshold", true)
	errThresholdTooSmall     = newHttpError(400, "ThresholdTooSmall", "Threshold too small", true)
	errThresholdTooBig       = newHttpError(400, "ThresholdTooBig", "Threshold too big", true)
	errThresholdTooPrecise   = newHttpError(400, "ThresholdTooPrecise", "Threshold too precise", true)
	errIncorrectIPAddress    = newHttpError(403, "IncorrectIPAddress", "Incorrect IP address", true)
	errRouteNotFound         = newHttpError(404, "RouteNotFound", "Route not found", false)
	errChainNotFound         = newHttpError(404, "ChainNotFound", "Chain not found", false)
	errPeriodNotFound        = newHttpError(404, "PeriodNotFound", "Period not found", false)
	errMetricNotFound        = newHttpError(404, "MetricNotFound", "Metric not found", false)
	errMinerNotFound         = newHttpError(404, "MinerNotFound", "Miner not found", false)
	errWorkerNotFound        = newHttpError(404, "WorkerNotFound", "Worker not found", false)
	errMethodNotAllowed      = newHttpError(405, "MethodNotAllowed", "Method not allowed", false)
	errInternalServerError   = newHttpError(500, "InternalServerError", "Internal server error", true)
)

type middleware func(*Context, http.Handler) http.Handler

type httpResponse struct {
	Status    int         `json:"status"`
	ErrorName *string     `json:"error"`
	Message   *string     `json:"message"`
	Data      interface{} `json:"data"`
	shouldLog bool
}

func (r httpResponse) Error() string {
	if r.ErrorName == nil || r.Message == nil {
		return ""
	}

	return fmt.Sprintf("%s (%d): %s", *r.ErrorName, r.Status, *r.Message)
}

func newHttpError(status int, name, msg string, shouldLog bool) httpResponse {
	res := httpResponse{
		Status:    status,
		ErrorName: &name,
		Message:   &msg,
		shouldLog: shouldLog,
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
