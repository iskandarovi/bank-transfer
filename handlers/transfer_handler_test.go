package handlers

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"bank-transfer/models"
	"bank-transfer/repositories"
)

type fakeTransferRepo struct {
	transferErr   error
	transferCalls int
}

func (f *fakeTransferRepo) Transfer(context.Context, int64, int64, float64) error {
	f.transferCalls++
	return f.transferErr
}

func (f *fakeTransferRepo) GetAll(context.Context) ([]models.Transfer, error) {
	return []models.Transfer{}, nil
}

func (f *fakeTransferRepo) GetByID(context.Context, int64) (models.Transfer, error) {
	return models.Transfer{}, nil
}

func TestTransferValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"valid request", `{"sender_id":1,"receiver_id":2,"amount":10.25}`, 200},
		{"same account", `{"sender_id":1,"receiver_id":1,"amount":10}`, 400},
		{"zero amount", `{"sender_id":1,"receiver_id":2,"amount":0}`, 400},
		{"negative ID", `{"sender_id":-1,"receiver_id":2,"amount":10}`, 400},
		{"more than two decimals", `{"sender_id":1,"receiver_id":2,"amount":1.234}`, 400},
		{"unknown field", `{"sender_id":1,"receiver_id":2,"amount":10,"extra":true}`, 400},
		{"invalid JSON", `not-json`, 400},
		{"two JSON values", `{"sender_id":1,"receiver_id":2,"amount":10} {}`, 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeTransferRepo{}
			handler := &TransferHandler{Repo: repo}
			request := httptest.NewRequest("POST", "/transfer", strings.NewReader(tt.body))
			response := httptest.NewRecorder()

			handler.Transfer(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, tt.wantStatus, response.Body.String())
			}
			if tt.wantStatus == 400 && repo.transferCalls != 0 {
				t.Fatal("noto'g'ri so'rov repositorygacha yetib bordi")
			}
		})
	}
}

func TestTransferRepositoryErrorsMapToHTTPStatus(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"account not found", repositories.ErrAccountNotFound, 404},
		{"insufficient balance", repositories.ErrInsufficientBalance, 400},
		{"invalid transfer", repositories.ErrInvalidTransfer, 400},
		{"database error", errors.New("database offline"), 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &TransferHandler{Repo: &fakeTransferRepo{transferErr: tt.err}}
			request := httptest.NewRequest("POST", "/transfer", strings.NewReader(`{"sender_id":1,"receiver_id":2,"amount":5}`))
			response := httptest.NewRecorder()

			handler.Transfer(response, request)
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
		})
	}
}
