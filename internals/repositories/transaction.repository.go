package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (tr *TransactionRepository) SoftDeleteTransaction(rctx context.Context, transactionId int) error {
	sql := `
		UPDATE transactions
		SET deleted_at = current_date
		WHERE id = $1
	`

	ctag, err := tr.db.Exec(rctx, sql, transactionId)
	if err != nil {
		return err
	}

	if ctag.RowsAffected() == 0 {
		return errors.New("no matching transaction id")
	}

	return nil
}
