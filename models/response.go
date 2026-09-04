package models

import "time"

type ValidationResponse struct {
	StatusCode        int                    `json:"status_code"`
	ErrorCode         string                 `json:"error_code,omitempty"`
	ParameterKey      string                 `json:"parameter_key"`
	RawParameterValue string                 `json:"raw_parameter_value"`
	ParameterValue    string                 `json:"parameter_value"`
	IsValid           bool                   `json:"is_valid"`
	Message           string                 `json:"message"`
	RequestID         string                 `json:"request_id"`
	Timestamp         time.Time              `json:"timestamp"`
	ExecutionTimeMs   int                    `json:"execution_time_ms"`
	LocationData      map[string]interface{} `json:"location_data,omitempty"`
	FromCache         bool                   `json:"from_cache"`
}

// BatchRequest is one form's worth of values, checked in a single call.
//
// The items are a list rather than an object keyed by document type: a form can
// legitimately carry two CPFs — a customer and a guarantor — and an object
// would silently keep only the last.
type BatchRequest struct {
	Items []BatchItem `json:"items"`
}

// BatchItem is one value to check.
type BatchItem struct {
	// ID is echoed back untouched so the caller can line results up with its
	// own form fields without relying on ordering.
	ID string `json:"id,omitempty"`
	// Key is the document type, matching the query parameter names of GET /validate.
	Key   string `json:"key"`
	Value string `json:"value"`
	// Qualifier carries the extra context a few documents need — today only the
	// issuing state for an inscrição estadual.
	Qualifier string `json:"qualifier,omitempty"`
}

// BatchResponse reports every item plus a count, so a caller can branch on the
// summary without walking the list.
type BatchResponse struct {
	StatusCode      int           `json:"status_code"`
	ErrorCode       string        `json:"error_code,omitempty"`
	Message         string        `json:"message,omitempty"`
	Summary         BatchSummary  `json:"summary"`
	Results         []BatchResult `json:"results"`
	RequestID       string        `json:"request_id"`
	Timestamp       time.Time     `json:"timestamp"`
	ExecutionTimeMs int           `json:"execution_time_ms"`
}

type BatchSummary struct {
	Total   int `json:"total"`
	Valid   int `json:"valid"`
	Invalid int `json:"invalid"`
}

type BatchResult struct {
	ID        string `json:"id,omitempty"`
	Key       string `json:"key"`
	ErrorCode string `json:"error_code,omitempty"`
	RawValue  string `json:"raw_value"`
	Value     string `json:"value"`
	IsValid   bool   `json:"is_valid"`
	Message   string `json:"message"`
	FromCache bool   `json:"from_cache"`
}
