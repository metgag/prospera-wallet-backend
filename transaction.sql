WITH user_participant AS (
    SELECT p.id AS participant_id
    FROM accounts a
    JOIN wallets w ON w.id = a.id
    JOIN participants p ON p.ref_id = w.id AND p.type = 'wallet'
    WHERE a.id = $1 -- ganti dengan user_id
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
ORDER BY d;

WITH user_participant AS (
    SELECT p.id AS participant_id
    FROM accounts a
    JOIN wallets w ON w.id = a.id
    JOIN participants p ON p.ref_id = w.id AND p.type = 'wallet'
    WHERE a.id = $1 -- ganti dengan user_id
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
ORDER BY week_start;
