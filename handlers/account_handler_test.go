package handlers

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"bank-transfer/models"

	"github.com/jackc/pgx/v5"
)

type fakeAccountRepo struct {
	account models.Account
	err     error
}

func (f *fakeAccountRepo) GetAll(context.Context) ([]models.Account, error) {
	return []models.Account{}, f.err
}

func (f *fakeAccountRepo) GetByID(context.Context, int64) (models.Account, error) {
	return f.account, f.err
}

func TestAccountGetByIDStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		err        error
		wantStatus int
	}{
		{"valid ID", "1", nil, 200},
		{"invalid ID", "abc", nil, 400},
		{"zero ID", "0", nil, 400},
		{"not found", "99", pgx.ErrNoRows, 404},
		{"database error", "1", errors.New("database offline"), 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &AccountHandler{Repo: &fakeAccountRepo{err: tt.err}}
			request := httptest.NewRequest("GET", "/accounts/"+tt.id, nil)
			request.SetPathValue("id", tt.id)
			response := httptest.NewRecorder()

			handler.GetByID(response, request)
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
		})
	}
}
