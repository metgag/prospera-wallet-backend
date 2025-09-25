package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Auth struct {
	db *pgxpool.Pool
}

func NewAuthRepo(db *pgxpool.Pool) *Auth {
	return &Auth{db: db}
}

func (r *Auth) Register(ctx context.Context, email, password string) error {
	// mulai transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction")
	}
	defer func() {
		// kalau ada panic / lupa commit → rollback
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// insert ke account
	var userID int
	queryAccount := `INSERT INTO accounts (email, password) VALUES ($1, $2) RETURNING id`
	if err = tx.QueryRow(ctx, queryAccount, email, password).Scan(&userID); err != nil {
		return fmt.Errorf("failed to insert accounts")
	}

	// insert ke profiles
	queryUser := `INSERT INTO profiles (id, fullname, phone, img) VALUES ($1, NULL, NULL, NULL);`
	if _, err = tx.Exec(ctx, queryUser, userID); err != nil {
		return fmt.Errorf("failed to insert profiles = %w", err)
	}

	// insert ke wallets
	queryWallet := `INSERT INTO wallets (id, balance) VALUES ($id, 0);`
	if _, err = tx.Exec(ctx, queryWallet, userID); err != nil {
		return fmt.Errorf("failed to insert wallets = %w", err)
	}

	// commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction")
	}

	return nil
}

func (r *Auth) Login(ctx context.Context, email string) (int, string, bool, error) {
	query := `SELECT id, password, pin FROM accounts WHERE email = $1`
	var id int
	var hashedPassword, pin string
	err := r.db.QueryRow(ctx, query, email).Scan(&id, &hashedPassword, &pin)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, "", false, nil
		}
		return 0, "", false, err
	}

	var isPinExist bool
	if pin != "" {
		isPinExist = true
	} else {
		isPinExist = false
	}

	return id, hashedPassword, isPinExist, nil
}
