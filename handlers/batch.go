package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"DataValidatorAPI/models"
	"DataValidatorAPI/utils"

	"github.com/google/uuid"
)

const (
	// maxBatchItems bounds the work one request can ask for. A form has tens of
	// fields, not thousands; anything past this is either a mistake or an
	// attempt to use the service as free compute.
	maxBatchItems = 100
	// maxBatchBytes bounds what will be read before parsing, so a large body
	// cannot exhaust memory ahead of the item check.
	maxBatchBytes = 1 << 20
)

// BatchHandler godoc
// @Summary Validates several values in one request
// @Description Checks a list of values, each with its own document type. Returns one result per item plus a summary.
// @Tags Validation
// @Accept json
// @Produce json
// @Param request body models.BatchRequest true "Values to validate"
// @Success 200 {object} models.BatchResponse
// @Failure 400 {object} models.BatchResponse
// @Router /validate/batch [post]
func BatchHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var request models.BatchRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBatchBytes))
	// Unknown fields are refused rather than ignored: a caller who misspells
	// "items" deserves to be told, not to receive an empty summary.
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		// The parser error is not echoed — it can quote the body, and the body
		// is exactly the personal data that must not travel back in an error.
		writeBatchError(w, start, "MALFORMED_BODY", "The request body could not be read as a batch of items")
		return
	}

	switch {
	case len(request.Items) == 0:
		writeBatchError(w, start, "EMPTY_BATCH", "No items to validate")
		return
	case len(request.Items) > maxBatchItems:
		writeBatchError(w, start, "BATCH_TOO_LARGE",
			"A batch carries at most "+strconv.Itoa(maxBatchItems)+" items")
		return
	}

	results := make([]models.BatchResult, 0, len(request.Items))
	summary := models.BatchSummary{Total: len(request.Items)}

	for _, item := range request.Items {
		result := validateItem(item)
		if result.IsValid {
			summary.Valid++
		} else {
			summary.Invalid++
		}
		results = append(results, result)
	}

	// A batch that ran is a successful request even when items failed
	// validation. The single-value endpoint answers 422 for an invalid value
	// because the value is the whole request; here it is one line of many, and
	// failing the response would discard the results that did pass.
	writeBatch(w, models.BatchResponse{
		StatusCode:      http.StatusOK,
		Summary:         summary,
		Results:         results,
		RequestID:       uuid.New().String(),
		Timestamp:       time.Now(),
		ExecutionTimeMs: int(time.Since(start).Milliseconds()),
	})
}

func validateItem(item models.BatchItem) models.BatchResult {
	key := strings.TrimSpace(strings.ToLower(item.Key))
	// "phone" is an accepted alias on the single endpoint and stays one here.
	if key == "phone" {
		key = "telephone"
	}

	result := models.BatchResult{ID: item.ID, Key: key, RawValue: item.Value, Value: item.Value}

	if strings.TrimSpace(item.Value) == "" {
		result.ErrorCode = "MISSING_VALUE"
		result.Message = "No value provided for this item"
		return result
	}

	outcome, known := utils.Validate(key, item.Value, item.Qualifier)
	if !known {
		// Distinguished from a failed validation on purpose: an unknown key is
		// the caller's mistake, and reporting it as "invalid" would suggest the
		// value was checked and rejected.
		result.ErrorCode = "UNKNOWN_KEY"
		result.Message = "Unknown document type " + key +
			" (expected one of " + strings.Join(utils.SupportedKeys(), ", ") + ")"
		return result
	}

	result.Value = outcome.Sanitized
	result.IsValid = outcome.Valid
	result.Message = outcome.Message
	result.FromCache = outcome.FromCache
	return result
}

func writeBatchError(w http.ResponseWriter, start time.Time, code, message string) {
	writeBatch(w, models.BatchResponse{
		StatusCode:      http.StatusBadRequest,
		ErrorCode:       code,
		Message:         message,
		Summary:         models.BatchSummary{},
		Results:         []models.BatchResult{},
		RequestID:       uuid.New().String(),
		Timestamp:       time.Now(),
		ExecutionTimeMs: int(time.Since(start).Milliseconds()),
	})
}

func writeBatch(w http.ResponseWriter, response models.BatchResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)
	_ = json.NewEncoder(w).Encode(response)
}
