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
// @Param pix query string false "PIX key in any of its five forms"
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

	if email := r.URL.Query().Get("email"); email != "" {
		isValid, sanitizedValue, message := utils.ValidateEmail(email)
		response = createResponse("email", email, sanitizedValue, isValid, message, start, false)
	} else if cpf := r.URL.Query().Get("cpf"); cpf != "" {
		isValid, sanitizedValue, message, fromCache := utils.ValidateCPFWithCache(cpf)
		response = createResponse("cpf", cpf, sanitizedValue, isValid, message, start, fromCache)
	} else if cnpj := r.URL.Query().Get("cnpj"); cnpj != "" {
		isValid, sanitizedValue, message := utils.ValidateCNPJ(cnpj)
		response = createResponse("cnpj", cnpj, sanitizedValue, isValid, message, start, false)
	} else if documento := r.URL.Query().Get("documento"); documento != "" {
		isValid, sanitizedValue, message := utils.ValidateDocument(documento)
		response = createResponse("documento", documento, sanitizedValue, isValid, message, start, false)
	} else if pis := r.URL.Query().Get("pis"); pis != "" {
		isValid, sanitizedValue, message := utils.ValidatePIS(pis)
		response = createResponse("pis", pis, sanitizedValue, isValid, message, start, false)
	} else if titulo := r.URL.Query().Get("titulo"); titulo != "" {
		isValid, sanitizedValue, message := utils.ValidateTituloEleitor(titulo)
		response = createResponse("titulo", titulo, sanitizedValue, isValid, message, start, false)
	} else if placa := r.URL.Query().Get("placa"); placa != "" {
		isValid, sanitizedValue, message := utils.ValidatePlate(placa)
		response = createResponse("placa", placa, sanitizedValue, isValid, message, start, false)
	} else if pix := r.URL.Query().Get("pix"); pix != "" {
		isValid, sanitizedValue, message := utils.ValidatePixKey(pix)
		response = createResponse("pix", pix, sanitizedValue, isValid, message, start, false)
	} else if name := r.URL.Query().Get("name"); name != "" {
		isValid, sanitizedValue, message := utils.ValidateName(name)
		response = createResponse("name", name, sanitizedValue, isValid, message, start, false)
	} else if telephone := firstNonEmpty(r.URL.Query().Get("telephone"), r.URL.Query().Get("phone")); telephone != "" {
		isValid, sanitizedValue, message := utils.ValidatePhone(telephone)
		response = createResponse("telephone", telephone, sanitizedValue, isValid, message, start, false)
	} else if plastic := r.URL.Query().Get("plastic"); plastic != "" {
		isValid, sanitizedValue, message := utils.ValidatePlastic(plastic)
		response = createResponse("plastic", plastic, sanitizedValue, isValid, message, start, false)
	} else if rg := r.URL.Query().Get("rg"); rg != "" {
		isValid, sanitizedValue, message := utils.ValidateRG(rg)
		response = createResponse("rg", rg, sanitizedValue, isValid, message, start, false)
	} else if cep := r.URL.Query().Get("cep"); cep != "" {
		isValid, sanitizedValue, message := utils.ValidateCEP(cep)
		response = createResponse("cep", cep, sanitizedValue, isValid, message, start, false)
	} else {
		response = models.ValidationResponse{
			StatusCode:      http.StatusBadRequest,
			ErrorCode:       "MISSING_PARAMETER",
			Message:         "No validation parameter provided",
			IsValid:         false,
			RequestID:       uuid.New().String(),
			Timestamp:       time.Now(),
			ExecutionTimeMs: int(time.Since(start).Milliseconds()),
		}
	}

	writeResponse(w, response)
}

func validationCount(r *http.Request) int {
	count := 0

	if strings.TrimSpace(r.URL.Query().Get("email")) != "" {
		count++
	}
	for _, key := range []string{"cnpj", "documento", "pis", "titulo", "placa", "pix"} {
		if strings.TrimSpace(r.URL.Query().Get(key)) != "" {
			count++
		}
	}
	if strings.TrimSpace(r.URL.Query().Get("cpf")) != "" {
		count++
	}
	if strings.TrimSpace(r.URL.Query().Get("name")) != "" {
		count++
	}
	if firstNonEmpty(r.URL.Query().Get("telephone"), r.URL.Query().Get("phone")) != "" {
		count++
	}
	if strings.TrimSpace(r.URL.Query().Get("plastic")) != "" {
		count++
	}
	if strings.TrimSpace(r.URL.Query().Get("rg")) != "" {
		count++
	}
	if strings.TrimSpace(r.URL.Query().Get("cep")) != "" {
		count++
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
