package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"bank-transfer/models"
	"bank-transfer/repositories"

	"github.com/jackc/pgx/v5"
)

type TransferReader interface {
	GetAll(ctx context.Context) ([]models.Transfer, error)
	GetByID(ctx context.Context, id int64) (models.Transfer, error)
}

type TransferWriter interface {
	Transfer(ctx context.Context, senderID, receiverID int64, amount float64) error
}

type TransferHandler struct {
	Repo interface {
		TransferReader
		TransferWriter
	}
}

type TransferRequest struct {
	SenderID   int64   `json:"sender_id"`
	ReceiverID int64   `json:"receiver_id"`
	Amount     float64 `json:"amount"`
}

func (h *TransferHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request TransferRequest
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "JSON noto'g'ri yoki kutilmagan maydon mavjud")
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "So'rovda faqat bitta JSON obyekt bo'lishi kerak")
		return
	}

	if request.SenderID <= 0 || request.ReceiverID <= 0 ||
		request.SenderID == request.ReceiverID {
		writeError(w, http.StatusBadRequest, "sender_id, receiver_id va amount qiymatlarini tekshiring; summa 2 xonali kasrgacha bo'lishi kerak")
		return
	}
	if _, validAmount := models.AmountToCents(request.Amount); !validAmount {
		writeError(w, http.StatusBadRequest, "amount 0 dan katta va 2 xonali kasr aniqligida bo'lishi kerak")
		return
	}

	err := h.Repo.Transfer(r.Context(), request.SenderID, request.ReceiverID, request.Amount)
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrInvalidTransfer):
			writeError(w, http.StatusBadRequest, "Transfer ma'lumotlari noto'g'ri")
		case errors.Is(err, repositories.ErrInsufficientBalance):
			writeError(w, http.StatusBadRequest, "Balans yetarli emas")
		case errors.Is(err, repositories.ErrAccountNotFound):
			writeError(w, http.StatusNotFound, "Jo'natuvchi yoki qabul qiluvchi hisobi topilmadi")
		default:
			log.Printf("POST /transfer: %v", err)
			writeError(w, http.StatusInternalServerError, "Pul o'tkazishda server xatosi")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Pul muvaffaqiyatli o'tkazildi"})
}

func (h *TransferHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	transfers, err := h.Repo.GetAll(r.Context())
	if err != nil {
		log.Printf("GET /transfers: %v", err)
		writeError(w, http.StatusInternalServerError, "Transferlarni olishda server xatosi")
		return
	}
	writeJSON(w, http.StatusOK, transfers)
}

func (h *TransferHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "Transfer ID musbat son bo'lishi kerak")
		return
	}

	transfer, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Transfer topilmadi")
			return
		}
		log.Printf("GET /transfers/%d: %v", id, err)
		writeError(w, http.StatusInternalServerError, "Transferni olishda server xatosi")
		return
	}
	writeJSON(w, http.StatusOK, transfer)
}
