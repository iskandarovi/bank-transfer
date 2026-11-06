package repositories

import (
	"context"
	"fmt"

	"bank-transfer/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepository struct {
	DB *pgxpool.Pool
}

func (r *AccountRepository) GetAll(ctx context.Context) ([]models.Account, error) {
	const query = `
		SELECT id, owner_name, balance, created_at
		FROM accounts
		ORDER BY id
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("accountlarni olishda xato: %w", err)
	}
	defer rows.Close()

	accounts := make([]models.Account, 0)
	for rows.Next() {
		var account models.Account
		if err := rows.Scan(
			&account.ID,
			&account.OwnerName,
			&account.Balance,
			&account.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("account ma'lumotlarini o'qishda xato: %w", err)
		}
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("account qatorlarini o'qishda xato: %w", err)
	}
	return accounts, nil
}

func (r *AccountRepository) GetByID(ctx context.Context, id int64) (models.Account, error) {
	const query = `
		SELECT id, owner_name, balance, created_at
		FROM accounts
		WHERE id = $1
	`

	var account models.Account
	err := r.DB.QueryRow(ctx, query, id).Scan(
		&account.ID,
		&account.OwnerName,
		&account.Balance,
		&account.CreatedAt,
	)
	if err != nil {
		return models.Account{}, fmt.Errorf("accountni olishda xato: %w", err)
	}
	return account, nil
}
