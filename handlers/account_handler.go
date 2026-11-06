package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"bank-transfer/models"

	"github.com/jackc/pgx/v5"
)

type AccountReader interface {
	GetAll(ctx context.Context) ([]models.Account, error)
	GetByID(ctx context.Context, id int64) (models.Account, error)
}

type AccountHandler struct {
	Repo AccountReader
}

func (h *AccountHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.Repo.GetAll(r.Context())
	if err != nil {
		log.Printf("GET /accounts: %v", err)
		writeError(w, http.StatusInternalServerError, "Accountlarni olishda server xatosi")
		return
	}
	writeJSON(w, http.StatusOK, accounts)
}

func (h *AccountHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "Account ID musbat son bo'lishi kerak")
		return
	}

	account, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Account topilmadi")
			return
		}
		log.Printf("GET /accounts/%d: %v", id, err)
		writeError(w, http.StatusInternalServerError, "Accountni olishda server xatosi")
		return
	}
	writeJSON(w, http.StatusOK, account)
}
