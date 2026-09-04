package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"DataValidatorAPI/models"
)

func TestValidateHandlerMissingParameter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/validate", nil)
	w := httptest.NewRecorder()

	ValidateHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var response models.ValidationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if response.ErrorCode != "MISSING_PARAMETER" {
		t.Fatalf("expected MISSING_PARAMETER, got %s", response.ErrorCode)
	}
}

func TestValidateHandlerMultipleParameters(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/validate?email=a%40a.com&cpf=52998224725", nil)
	w := httptest.NewRecorder()

	ValidateHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestValidateHandlerValidationFailed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/validate?email=invalid", nil)
	w := httptest.NewRecorder()

	ValidateHandler(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
}

func TestValidateHandlerPhoneAlias(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/validate?phone=11912345678", nil)
	w := httptest.NewRecorder()

	ValidateHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var response models.ValidationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if response.ParameterKey != "telephone" {
		t.Fatalf("expected parameter_key telephone, got %s", response.ParameterKey)
	}
}

func TestValidateHandlerPhoneAndTelephoneAliasTogether(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/validate?phone=11912345678&telephone=11912345678", nil)
	w := httptest.NewRecorder()

	ValidateHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestValidateHandlerCPFValidationFailureStatus(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/validate?cpf=11111111111", nil)
	w := httptest.NewRecorder()

	ValidateHandler(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}

	var response models.ValidationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if response.ErrorCode != "VALIDATION_FAILED" {
		t.Fatalf("expected VALIDATION_FAILED, got %s", response.ErrorCode)
	}
}

func TestValidateHandlerCNPJ(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantValid  bool
	}{
		{"valid_cnpj", "cnpj=33.000.167/0001-01", http.StatusOK, true},
		{"invalid_cnpj", "cnpj=33.000.167/0001-11", http.StatusUnprocessableEntity, false},
		// CNPJ has to take part in the one-parameter-per-request rule, or it
		// becomes the hole through which two validations arrive at once.
		{"cnpj_with_cpf", "cnpj=33.000.167/0001-01&cpf=529.982.247-25", http.StatusBadRequest, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ValidateHandler(recorder, httptest.NewRequest(http.MethodGet, "/validate?"+tt.query, nil))

			if recorder.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tt.wantStatus, recorder.Code, recorder.Body.String())
			}

			var response models.ValidationResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("could not decode the response: %v", err)
			}
			if response.IsValid != tt.wantValid {
				t.Fatalf("expected is_valid=%v, got %v (%s)", tt.wantValid, response.IsValid, response.Message)
			}
		})
	}
}
