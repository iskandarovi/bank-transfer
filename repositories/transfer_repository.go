package repositories

import (
	"context"
	"errors"
	"fmt"

	"bank-transfer/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidTransfer     = errors.New("transfer ma'lumotlari noto'g'ri")
	ErrAccountNotFound     = errors.New("hisob topilmadi")
	ErrInsufficientBalance = errors.New("balans yetarli emas")
)

type TransferRepository struct {
	DB *pgxpool.Pool
}

func (r *TransferRepository) Transfer(ctx context.Context, senderID, receiverID int64, amount float64) error {
	amountCents, validAmount := models.AmountToCents(amount)
	if senderID <= 0 || receiverID <= 0 || senderID == receiverID || !validAmount {
		return ErrInvalidTransfer
	}
	amountDecimal := fmt.Sprintf("%d.%02d", amountCents/100, amountCents%100)

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("transaction boshlanmadi: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Hisoblarni bir xil tartibda bloklash deadlock ehtimolini kamaytiradi.
	rows, err := tx.Query(ctx, `
		SELECT id, (balance * 100)::bigint
		FROM accounts
		WHERE id IN ($1, $2)
		ORDER BY id
		FOR UPDATE
	`, senderID, receiverID)
	if err != nil {
		return fmt.Errorf("hisoblarni tekshirishda xato: %w", err)
	}

	balances := make(map[int64]int64, 2)
	for rows.Next() {
		var id int64
		var balanceCents int64
		if err := rows.Scan(&id, &balanceCents); err != nil {
			rows.Close()
			return fmt.Errorf("hisob balansini o'qishda xato: %w", err)
		}
		balances[id] = balanceCents
	}
	rowsErr := rows.Err()
	rows.Close()
	if rowsErr != nil {
		return fmt.Errorf("hisob qatorlarini o'qishda xato: %w", rowsErr)
	}

	senderBalanceCents, senderExists := balances[senderID]
	_, receiverExists := balances[receiverID]
	if !senderExists || !receiverExists {
		return ErrAccountNotFound
	}
	if senderBalanceCents < amountCents {
		return ErrInsufficientBalance
	}

	if _, err := tx.Exec(ctx, `
		UPDATE accounts SET balance = balance - $1::numeric WHERE id = $2
	`, amountDecimal, senderID); err != nil {
		return fmt.Errorf("jo'natuvchi balansini yangilashda xato: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE accounts SET balance = balance + $1::numeric WHERE id = $2
	`, amountDecimal, receiverID); err != nil {
		return fmt.Errorf("qabul qiluvchi balansini yangilashda xato: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO transfers (sender_id, receiver_id, amount)
		VALUES ($1, $2, $3::numeric)
	`, senderID, receiverID, amountDecimal); err != nil {
		return fmt.Errorf("transfer tarixini saqlashda xato: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("transaction tasdiqlanmadi: %w", err)
	}
	return nil
}

func (r *TransferRepository) GetAll(ctx context.Context) ([]models.Transfer, error) {
	const query = `
		SELECT t.id, t.sender_id, sender.owner_name,
		       t.receiver_id, receiver.owner_name, t.amount, t.created_at
		FROM transfers t
		JOIN accounts sender ON sender.id = t.sender_id
		JOIN accounts receiver ON receiver.id = t.receiver_id
		ORDER BY t.id DESC
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("transferlarni olishda xato: %w", err)
	}
	defer rows.Close()

	transfers := make([]models.Transfer, 0)
	for rows.Next() {
		var transfer models.Transfer
		if err := rows.Scan(
			&transfer.ID,
			&transfer.SenderID,
			&transfer.SenderName,
			&transfer.ReceiverID,
			&transfer.ReceiverName,
			&transfer.Amount,
			&transfer.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("transfer ma'lumotlarini o'qishda xato: %w", err)
		}
		transfers = append(transfers, transfer)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("transfer qatorlarini o'qishda xato: %w", err)
	}
	return transfers, nil
}

func (r *TransferRepository) GetByID(ctx context.Context, id int64) (models.Transfer, error) {
	const query = `
		SELECT t.id, t.sender_id, sender.owner_name,
		       t.receiver_id, receiver.owner_name, t.amount, t.created_at
		FROM transfers t
		JOIN accounts sender ON sender.id = t.sender_id
		JOIN accounts receiver ON receiver.id = t.receiver_id
		WHERE t.id = $1
	`

	var transfer models.Transfer
	err := r.DB.QueryRow(ctx, query, id).Scan(
		&transfer.ID,
		&transfer.SenderID,
		&transfer.SenderName,
		&transfer.ReceiverID,
		&transfer.ReceiverName,
		&transfer.Amount,
		&transfer.CreatedAt,
	)
	if err != nil {
		return models.Transfer{}, fmt.Errorf("transferni olishda xato: %w", err)
	}
	return transfer, nil
}
