package repositories

import (
	"context"
	"fmt"

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
	queryUser := `INSERT INTO profiles (fullname, phone, img) VALUES (NULL, NULL, NULL);`
	if _, err = tx.Exec(ctx, queryUser); err != nil {
		return fmt.Errorf("failed to insert profiles = %w", err)
	}

	// insert ke wallets
	queryWallet := `INSERT INTO wallets (balance) VALUES (0);`
	if _, err = tx.Exec(ctx, queryWallet); err != nil {
		return fmt.Errorf("failed to insert wallets = %w", err)
	}

	// commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction")
	}

	return nil
}

// func (r *Auth) Login(ctx context.Context, email string) (int, string, string, error) {
// 	query := `SELECT id, role, password FROM account WHERE email = $1`
// 	var id int
// 	var role string
// 	var hashedPassword string

// 	err := r.db.QueryRow(ctx, query, email).Scan(&id, &role, &hashedPassword)
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			return 0, "", "", nil // ✅ biar bisa bedain di handler
// 		}
// 		return 0, "", "", err
// 	}
// 	return id, role, hashedPassword, nil
// }

// func (r *Auth) BlacklistToken(ctx context.Context, token string, expiresIn time.Duration) error {
// 	data := models.BlacklistToken{
// 		Token:     token,
// 		ExpiresIn: expiresIn,
// 	}
// 	bt, err := json.Marshal(data)
// 	if err != nil {
// 		log.Println("❌ Internal Server Error.\nCause:", err)
// 		return err
// 	}

// 	if err := r.rdb.Set(ctx, "blacklist:"+token, bt, expiresIn).Err(); err != nil {
// 		log.Printf("❌ Redis Error.\nCause: %s\n", err)
// 		return err
// 	}

// 	return nil
// }
