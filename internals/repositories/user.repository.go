package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prospera/internals/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// GET PROFILE
func (ur *UserRepository) GetProfile(ctx context.Context, uid int) (*models.Profile, error) {
	sql := `
		SELECT fullname, phone, img, verified
		FROM profiles
		WHERE id = $1
	`

	row := ur.db.QueryRow(ctx, sql, uid)

	var profile models.Profile
	err := row.Scan(
		&profile.FullName,
		&profile.PhoneNumber,
		&profile.Avatar,
		&profile.Verified,
	)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// UPDATE PROFILE
func (ur *UserRepository) UpdateProfile(ctx context.Context, uid int, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	setClauses := []string{}
	args := []any{}
	i := 1

	for col, val := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	// cek apakah ada fullname, phone, img yang diupdate
	needVerify := false
	if _, ok := updates["fullname"]; ok {
		needVerify = true
	}
	if _, ok := updates["phone"]; ok {
		needVerify = true
	}
	if _, ok := updates["img"]; ok {
		needVerify = true
	}

	// kalau fullname, phone, img semuanya sudah terisi → verified = true
	if needVerify {
		setClauses = append(setClauses,
			"verified = (CASE WHEN fullname IS NOT NULL AND fullname <> '' AND phone IS NOT NULL AND phone <> '' AND img IS NOT NULL AND img <> '' THEN TRUE ELSE verified END)",
		)
	}

	// tambah updated_at
	setClauses = append(setClauses, "updated_at = NOW()")

	query := fmt.Sprintf(`
		UPDATE profiles
		SET %s
		WHERE id = $%d
	`, strings.Join(setClauses, ", "), i)

	args = append(args, uid)

	_, err := ur.db.Exec(ctx, query, args...)
	return err
}

// GET ALL USERS
func (ur *UserRepository) GetAllUser(rctx context.Context, uid int) ([]models.User, error) {
	sql := `
		SELECT fullname, phone, img
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
			&user.PhoneNumber,
			&user.Avatar,
		); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

// GET HISTORY TRANSACTIONS
func (ur *UserRepository) GetUserHistoryTransactions(ctx context.Context, userID, limit, offset int) ([]models.TransactionHistory, int, error) {
	// Hitung total data untuk pagination
	countQuery := `
		WITH user_participant AS (
			SELECT p.id AS participant_id
			FROM accounts a
			JOIN wallets w ON w.id = a.id
			JOIN participants p ON p.ref_id = w.id AND p.type = 'wallet'
			WHERE a.id = $1
		)
		SELECT COUNT(*)
		FROM transactions t
		JOIN user_participant up 
			ON t.id_sender = up.participant_id OR t.id_receiver = up.participant_id
	`
	var total int
	err := ur.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Query utama + limit offset
	query := `
	WITH user_participant AS (
		SELECT p.id AS participant_id
		FROM accounts a
		JOIN wallets w ON w.id = a.id
		JOIN participants p ON p.ref_id = w.id AND p.type = 'wallet'
		WHERE a.id = $1
	)
	SELECT 
		t.id,
		t.type,
		t.total,
		CASE 
			WHEN t.id_sender = up.participant_id THEN 'debit'
			WHEN t.id_receiver = up.participant_id THEN 'credit'
		END AS direction,
		CASE 
			WHEN t.id_sender = up.participant_id THEN pr.type
			ELSE ps.type
		END AS counterparty_type,
		COALESCE(
			(CASE 
				WHEN t.id_sender = up.participant_id AND pr.type = 'wallet' THEN prf.fullname
				WHEN t.id_receiver = up.participant_id AND ps.type = 'wallet' THEN prf.fullname
			END),
			(CASE 
				WHEN t.id_sender = up.participant_id AND pr.type = 'internal' THEN ia.name
				WHEN t.id_receiver = up.participant_id AND ps.type = 'internal' THEN ia.name
			END)
		) AS counterparty_name,
		COALESCE(
			(CASE 
				WHEN t.id_sender = up.participant_id AND pr.type = 'wallet' THEN prf.img
				WHEN t.id_receiver = up.participant_id AND ps.type = 'wallet' THEN prf.img
			END),
			(CASE 
				WHEN t.id_sender = up.participant_id AND pr.type = 'internal' THEN ia.img
				WHEN t.id_receiver = up.participant_id AND ps.type = 'internal' THEN ia.img
			END)
		) AS counterparty_img,
		CASE 
			WHEN (t.id_sender = up.participant_id AND pr.type = 'wallet') THEN prf.phone
			WHEN (t.id_receiver = up.participant_id AND ps.type = 'wallet') THEN prf.phone
			ELSE NULL
		END AS counterparty_phone,
		t.created_at
	FROM transactions t
	JOIN user_participant up 
		ON t.id_sender = up.participant_id OR t.id_receiver = up.participant_id
	JOIN participants ps ON ps.id = t.id_sender
	JOIN participants pr ON pr.id = t.id_receiver
	LEFT JOIN wallets w 
		ON ( (ps.type = 'wallet' AND ps.ref_id = w.id) OR (pr.type = 'wallet' AND pr.ref_id = w.id) )
	LEFT JOIN profiles prf ON prf.id = w.id
	LEFT JOIN internal_accounts ia 
		ON ( (ps.type = 'internal' AND ps.ref_id = ia.id) OR (pr.type = 'internal' AND pr.ref_id = ia.id) )
	ORDER BY DATE(t.created_at) DESC, t.created_at DESC
	LIMIT $2 OFFSET $3;
	`

	rows, err := ur.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []models.TransactionHistory
	for rows.Next() {
		var tx models.TransactionHistory
		err := rows.Scan(
			&tx.ID,
			&tx.Type,
			&tx.Total,
			&tx.Direction,
			&tx.CounterpartyType,
			&tx.CounterpartyName,
			&tx.CounterpartyImg,
			&tx.CounterpartyPhone,
			&tx.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, total, nil
}

// DELETE HISTORY TRANSACTIONS
func (ur *UserRepository) SoftDeleteTransaction(rctx context.Context, uid, transactionId int) error {
	sql := `
		UPDATE transactions SET 
		deleted_sender =
		CASE
		WHEN id_sender = $1 THEN CURRENT_DATE
		ELSE deleted_sender
		END, 
		deleted_receiver =
		CASE
		WHEN id_receiver = $1 THEN CURRENT_DATE
		ELSE deleted_receiver
		END
		WHERE id = $2
	`

	ctag, err := ur.db.Exec(rctx, sql, uid, transactionId)
	if err != nil {
		return err
	}

	if ctag.RowsAffected() == 0 {
		return errors.New("no matching transaction id")
	}

	return nil
}

// PATCH CHANGE PASSWORD
func (ur *UserRepository) GetPasswordFromID(ctx context.Context, id int) (string, error) {
	query := `
		SELECT
			password
		FROM
			accounts
		WHERE
			id = $1`

	var userPass string
	if err := ur.db.QueryRow(ctx, query, id).Scan(&userPass); err != nil {
		return "", errors.New("failed to get password")
	}
	return userPass, nil
}

func (ur *UserRepository) ChangePassword(ctx context.Context, userID int, hashedPassword string) error {
	query := `
		UPDATE
			accounts
		SET
			password = $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE
			id = $2`
	_, err := ur.db.Exec(ctx, query, hashedPassword, userID)
	if err != nil {
		return err
	}
	return nil
}

// POST SUMMARY MONTH AND WEEK
func (r *UserRepository) GetDailySummary(ctx context.Context, userID int) ([]models.DailySummary, error) {
	query := `
	WITH user_participant AS (
		SELECT p.id AS participant_id
		FROM accounts a
		JOIN wallets w ON w.id = a.id
		JOIN participants p ON p.ref_id = w.id AND p.type = 'wallet'
		WHERE a.id = $1
	),
	daily_summary AS (
		SELECT
			DATE(t.created_at) AS date,
			SUM(CASE WHEN t.id_sender = up.participant_id THEN t.total ELSE 0 END) AS total_expense,
			SUM(CASE WHEN t.id_receiver = up.participant_id THEN t.total ELSE 0 END) AS total_income
		FROM transactions t
		JOIN user_participant up 
			ON t.id_sender = up.participant_id OR t.id_receiver = up.participant_id
		WHERE t.created_at >= CURRENT_DATE - interval '6 days'
		GROUP BY DATE(t.created_at)
	)
	SELECT 
		d::date AS date,
		COALESCE(ds.total_expense, 0) AS total_expense,
		COALESCE(ds.total_income, 0) AS total_income
	FROM generate_series(
		CURRENT_DATE - interval '6 days',
		CURRENT_DATE,
		interval '1 day'
	) d
	LEFT JOIN daily_summary ds ON ds.date = d::date
	ORDER BY d;`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []models.DailySummary
	for rows.Next() {
		var s models.DailySummary
		if err := rows.Scan(&s.Date, &s.TotalExpense, &s.TotalIncome); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}

	return summaries, nil
}

func (r *UserRepository) GetWeeklySummary(ctx context.Context, userID int) ([]models.WeeklySummary, error) {
	query := `
	WITH user_participant AS (
		SELECT p.id AS participant_id
		FROM accounts a
		JOIN wallets w ON w.id = a.id
		JOIN participants p ON p.ref_id = w.id AND p.type = 'wallet'
		WHERE a.id = $1
	),
	weekly_summary AS (
		SELECT
			date_trunc('week', t.created_at)::date AS week_start,
			SUM(CASE WHEN t.id_sender = up.participant_id THEN t.total ELSE 0 END) AS total_expense,
			SUM(CASE WHEN t.id_receiver = up.participant_id THEN t.total ELSE 0 END) AS total_income
		FROM transactions t
		JOIN user_participant up 
			ON t.id_sender = up.participant_id OR t.id_receiver = up.participant_id
		WHERE t.created_at >= date_trunc('month', CURRENT_DATE)
		AND t.created_at < (date_trunc('month', CURRENT_DATE) + interval '1 month')
		GROUP BY date_trunc('week', t.created_at)
	)
	SELECT 
		week_start,
		week_start + interval '6 days' AS week_end,
		COALESCE(ws.total_expense, 0) AS total_expense,
		COALESCE(ws.total_income, 0) AS total_income
	FROM weekly_summary ws
	ORDER BY week_start;`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []models.WeeklySummary
	for rows.Next() {
		var s models.WeeklySummary
		if err := rows.Scan(&s.WeekStart, &s.WeekEnd, &s.TotalExpense, &s.TotalIncome); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}

	return summaries, nil
}
