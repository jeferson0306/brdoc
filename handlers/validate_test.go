package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jeferson0306/api-data-validator/models"
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

func TestValidateHandlerNewDocuments(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantValid  bool
		wantKey    string
	}{
		{"documento_cpf", "documento=529.982.247-25", http.StatusOK, true, "documento"},
		{"documento_cnpj", "documento=33.000.167/0001-01", http.StatusOK, true, "documento"},
		{"pis", "pis=12056874107", http.StatusOK, true, "pis"},
		{"pis_invalid", "pis=12056412547", http.StatusUnprocessableEntity, false, "pis"},
		{"placa_mercosul", "placa=ABC1D23", http.StatusOK, true, "placa"},
		{"placa_invalid", "placa=AB1CD23", http.StatusUnprocessableEntity, false, "placa"},
		{"pix_email", "pix=jeferson@example.com", http.StatusOK, true, "pix"},
		{"pix_invalid", "pix=minha-chave", http.StatusUnprocessableEntity, false, "pix"},
		// Each new parameter must join the one-per-request rule, or it becomes
		// a way to smuggle a second validation past the guard.
		{"two_new_parameters", "pis=12056874107&placa=ABC1D23", http.StatusBadRequest, false, ""},
		{"new_parameter_with_cpf", "pix=jeferson@example.com&cpf=529.982.247-25", http.StatusBadRequest, false, ""},
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
			if tt.wantKey != "" && response.ParameterKey != tt.wantKey {
				t.Fatalf("expected parameter_key %q, got %q", tt.wantKey, response.ParameterKey)
			}
		})
	}
}

func TestValidateHandlerInscricaoEstadual(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantValid  bool
	}{
		{"valid_with_state", "ie=0100482300112&uf=AC", http.StatusOK, true},
		{"valid_elsewhere", "ie=0100482300112&uf=SP", http.StatusUnprocessableEntity, false},
		{"missing_state", "ie=0100482300112", http.StatusUnprocessableEntity, false},
		{"unknown_state", "ie=0100482300112&uf=XX", http.StatusUnprocessableEntity, false},
		{"exempt", "ie=ISENTO&uf=SP", http.StatusOK, true},
		// uf qualifies ie rather than being a validation of its own, so it must
		// not push the request over the one-parameter limit.
		{"uf_is_not_a_second_validation", "ie=0100482300112&uf=AC", http.StatusOK, true},
		{"ie_with_cpf", "ie=0100482300112&uf=AC&cpf=529.982.247-25", http.StatusBadRequest, false},
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

func TestValidateHandlerVehicleAndBoleto(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantValid  bool
	}{
		{"cnh", "cnh=00000000119", http.StatusOK, true},
		{"cnh_invalid", "cnh=12345678901", http.StatusUnprocessableEntity, false},
		{"renavam_nine", "renavam=639884962", http.StatusOK, true},
		{"renavam_eleven", "renavam=00639884962", http.StatusOK, true},
		{"boleto", "boleto=00190000090114971860168524522114675860000102656", http.StatusOK, true},
		{"boleto_arrecadacao_is_not_checked", "boleto=846700000017435900240200610207807116000000000000", http.StatusUnprocessableEntity, false},
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

// The batch endpoint reads the same table, so the three new documents must be
// reachable there without another line of wiring.
func TestBatchReachesTheNewDocuments(t *testing.T) {
	recorder, response := postBatch(t, `{"items":[
		{"key":"cnh","value":"00000000119"},
		{"key":"renavam","value":"639884962"},
		{"key":"boleto","value":"00190000090114971860168524522114675860000102656"}
	]}`)

	if recorder.Code != http.StatusOK || response.Summary.Valid != 3 {
		t.Fatalf("expected all three to validate through the batch: %+v", response)
	}
}
