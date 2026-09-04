package handlers

import (
	"DataValidatorAPI/models"
	"DataValidatorAPI/utils"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ValidateHandler godoc
// @Summary Validates provided data format
// @Description Validates different types of data, including email, CPF, name, phone, RG, CEP, and credit card number
// @Tags Validation
// @Param email query string false "Email to be validated"
// @Param cpf query string false "CPF to be validated"
// @Param cnpj query string false "CNPJ to be validated"
// @Param documento query string false "CPF or CNPJ, whichever the digits describe"
// @Param pis query string false "PIS/PASEP/NIT/NIS to be validated"
// @Param titulo query string false "Voter registration number to be validated"
// @Param placa query string false "Vehicle plate, old or Mercosul pattern"
// @Param cnh query string false "Driving licence number to be validated"
// @Param renavam query string false "Vehicle registration number, 9 or 11 digits"
// @Param boleto query string false "Bank slip linha digitável, 47 digits"
// @Param pix query string false "PIX key in any of its five forms"
// @Param ie query string false "State registration; requires uf"
// @Param uf query string false "Two-letter state code qualifying ie"
// @Param name query string false "Name to be validated"
// @Param telephone query string false "Phone number to be validated"
// @Param phone query string false "Phone number to be validated"
// @Param plastic query string false "Credit card number to be validated"
// @Param rg query string false "RG to be validated"
// @Param cep query string false "CEP (postal code) to be validated"
// @Success 200 {object} models.ValidationResponse
// @Failure 400 {object} models.ValidationResponse
// @Failure 422 {object} models.ValidationResponse
// @Router /validate [get]
func ValidateHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var response models.ValidationResponse

	if validationCount(r) > 1 {
		response = models.ValidationResponse{
			StatusCode:      http.StatusBadRequest,
			ErrorCode:       "MULTIPLE_PARAMETERS",
			Message:         "Only one validation parameter is allowed per request",
			IsValid:         false,
			RequestID:       uuid.New().String(),
			Timestamp:       time.Now(),
			ExecutionTimeMs: int(time.Since(start).Milliseconds()),
		}
		writeResponse(w, response)
		return
	}

	key, value := firstProvided(r)
	if key == "" {
		writeResponse(w, models.ValidationResponse{
			StatusCode:      http.StatusBadRequest,
			ErrorCode:       "MISSING_PARAMETER",
			Message:         "No validation parameter provided",
			IsValid:         false,
			RequestID:       uuid.New().String(),
			Timestamp:       time.Now(),
			ExecutionTimeMs: int(time.Since(start).Milliseconds()),
		})
		return
	}

	// The qualifier is only read for the validations that take one; passing it
	// unconditionally keeps this loop from growing a special case per document.
	outcome, _ := utils.Validate(key, value, r.URL.Query().Get("uf"))
	response = createResponse(key, value, outcome.Sanitized, outcome.Valid, outcome.Message, start, outcome.FromCache)

	writeResponse(w, response)
}

// firstProvided returns the first validation parameter present, in the order
// utils.ValidationKeys documents. "phone" remains an accepted alias for
// "telephone" because it was published that way.
func firstProvided(r *http.Request) (string, string) {
	for _, key := range utils.ValidationKeys {
		value := r.URL.Query().Get(key)
		if key == "telephone" {
			value = firstNonEmpty(value, r.URL.Query().Get("phone"))
		}
		if strings.TrimSpace(value) != "" {
			return key, value
		}
	}
	return "", ""
}

func validationCount(r *http.Request) int {
	count := 0

	for _, key := range utils.ValidationKeys {
		value := r.URL.Query().Get(key)
		if key == "telephone" {
			value = firstNonEmpty(value, r.URL.Query().Get("phone"))
		}
		if strings.TrimSpace(value) != "" {
			count++
		}
	}

	return count
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func writeResponse(w http.ResponseWriter, response models.ValidationResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)
	json.NewEncoder(w).Encode(response)
}

// createResponse creates a response with the provided data.
func createResponse(key, rawValue, sanitizedValue string, isValid bool, message string, start time.Time, fromCache bool) models.ValidationResponse {
	statusCode := http.StatusOK
	errorCode := ""
	if !isValid {
		statusCode = http.StatusUnprocessableEntity
		errorCode = "VALIDATION_FAILED"
	}

	return models.ValidationResponse{
		StatusCode:        statusCode,
		ErrorCode:         errorCode,
		ParameterKey:      key,
		RawParameterValue: rawValue,
		ParameterValue:    sanitizedValue,
		IsValid:           isValid,
		Message:           message,
		RequestID:         uuid.New().String(),
		Timestamp:         time.Now(),
		ExecutionTimeMs:   int(time.Since(start).Milliseconds()),
		FromCache:         fromCache,
	}
}
