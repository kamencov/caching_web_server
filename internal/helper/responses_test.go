package helper

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOkResponse(t *testing.T) {
	w := httptest.NewRecorder()
	OkResponse(w, "test")

	if w.Code != http.StatusOK {
		t.Errorf("Status code is not correct. Got %d, want %d.", w.Code, http.StatusOK)
	}
}

func TestFailResponse(t *testing.T) {
	w := httptest.NewRecorder()
	FailResponse(w, http.StatusBadRequest, "test")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code is not correct. Got %d, want %d.", w.Code, http.StatusBadRequest)
	}
}

func TestOkDataResponse(t *testing.T) {
	w := httptest.NewRecorder()
	OkDataResponse(w, "test")

	if w.Code != http.StatusOK {
		t.Errorf("Status code is not correct. Got %d, want %d.", w.Code, http.StatusOK)
	}
}
