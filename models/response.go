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
