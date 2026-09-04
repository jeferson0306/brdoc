package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"DataValidatorAPI/models"
)

func postBatch(t *testing.T, body string) (*httptest.ResponseRecorder, models.BatchResponse) {
	t.Helper()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/validate/batch", strings.NewReader(body))
	BatchHandler(recorder, request)

	var response models.BatchResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("could not decode the response: %v — body was %s", err, recorder.Body.String())
	}
	return recorder, response
}

func TestBatchValidatesEveryItem(t *testing.T) {
	recorder, response := postBatch(t, `{"items":[
		{"id":"a","key":"cpf","value":"529.982.247-25"},
		{"id":"b","key":"cnpj","value":"33.000.167/0001-01"},
		{"id":"c","key":"email","value":"jeferson@example.com"},
		{"id":"d","key":"ie","value":"0100482300112","qualifier":"AC"},
		{"id":"e","key":"cpf","value":"529.982.247-26"}
	]}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if response.Summary.Total != 5 || response.Summary.Valid != 4 || response.Summary.Invalid != 1 {
		t.Fatalf("unexpected summary: %+v", response.Summary)
	}
	// The ids come back untouched, so a caller can line results up with its own
	// fields without depending on ordering.
	for i, want := range []string{"a", "b", "c", "d", "e"} {
		if response.Results[i].ID != want {
			t.Fatalf("result %d carried id %q, expected %q", i, response.Results[i].ID, want)
		}
	}
	if !response.Results[0].IsValid || response.Results[4].IsValid {
		t.Fatalf("the CPFs were judged wrongly: %+v", response.Results)
	}
}

// A form can legitimately carry two of the same document — a customer and a
// guarantor — which is why items are a list rather than an object.
func TestBatchAcceptsRepeatedKeys(t *testing.T) {
	_, response := postBatch(t, `{"items":[
		{"id":"titular","key":"cpf","value":"529.982.247-25"},
		{"id":"avalista","key":"cpf","value":"168.995.350-09"}
	]}`)

	if response.Summary.Total != 2 || response.Summary.Valid != 2 {
		t.Fatalf("expected both CPFs to be checked and valid: %+v", response.Summary)
	}
}

// One bad item must not discard the results that did pass.
func TestBatchReportsUnknownKeyWithoutFailingTheRequest(t *testing.T) {
	recorder, response := postBatch(t, `{"items":[
		{"key":"cpf","value":"529.982.247-25"},
		{"key":"passaporte","value":"AB123456"}
	]}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected the batch itself to succeed, got %d", recorder.Code)
	}
	if response.Results[1].ErrorCode != "UNKNOWN_KEY" {
		t.Fatalf("expected UNKNOWN_KEY, got %+v", response.Results[1])
	}
	if !response.Results[0].IsValid {
		t.Fatal("a valid item was lost because another item had a bad key")
	}
}

func TestBatchRejectsBadRequests(t *testing.T) {
	tests := []struct {
		name string
		body string
		code string
	}{
		{"malformed_json", `{"items":[`, "MALFORMED_BODY"},
		{"unknown_field", `{"itens":[{"key":"cpf","value":"1"}]}`, "MALFORMED_BODY"},
		{"empty", `{"items":[]}`, "EMPTY_BATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder, response := postBatch(t, tt.body)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", recorder.Code)
			}
			if response.ErrorCode != tt.code {
				t.Fatalf("expected %s, got %s", tt.code, response.ErrorCode)
			}
		})
	}
}

func TestBatchRefusesOversizedBatches(t *testing.T) {
	items := make([]string, 0, maxBatchItems+1)
	for i := 0; i <= maxBatchItems; i++ {
		items = append(items, `{"key":"cpf","value":"529.982.247-25"}`)
	}

	recorder, response := postBatch(t, fmt.Sprintf(`{"items":[%s]}`, strings.Join(items, ",")))
	if recorder.Code != http.StatusBadRequest || response.ErrorCode != "BATCH_TOO_LARGE" {
		t.Fatalf("expected BATCH_TOO_LARGE, got %d %s", recorder.Code, response.ErrorCode)
	}
}

// A parser error can quote the body, and the body is the personal data that
// must never travel back out in an error message.
func TestBatchErrorDoesNotEchoTheBody(t *testing.T) {
	_, response := postBatch(t, `{"items":[{"key":"cpf","value":"529.982.247-25"}`)

	if strings.Contains(response.Message, "529") || strings.Contains(response.Message, "982") {
		t.Fatalf("the error message carries the submitted value: %s", response.Message)
	}
}

func TestBatchMissingValue(t *testing.T) {
	_, response := postBatch(t, `{"items":[{"key":"cpf","value":"   "}]}`)

	if response.Results[0].ErrorCode != "MISSING_VALUE" {
		t.Fatalf("expected MISSING_VALUE, got %+v", response.Results[0])
	}
}
