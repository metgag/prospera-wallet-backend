package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prospera/internals/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (ur *UserRepository) GetUser(rctx context.Context, uid int) ([]models.User, error) {
	sql := `
		SELECT fullname
		FROM profiles
		WHERE id != $1
	`

	rows, err := ur.db.Query(rctx, sql, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.FullName,
		); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}
