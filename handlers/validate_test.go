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
