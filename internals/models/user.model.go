package models

import "time"

type Profile struct {
	ID          int     `json:"id"`
	FullName    *string `json:"full_name"`
	PhoneNumber *string `json:"phone_number"`
	Avatar      *string `json:"avatar"`
	Verified    bool    `json:"verified"`
	Email       string  `json:"email"`
}

type User struct {
	ID          int     `json:"id"`
	FullName    *string `json:"full_name"`
	PhoneNumber *string `json:"phone_number"`
	Avatar      *string `json:"avatar"`
	Verified    bool    `json:"verified"`
}

type TransactionHistory struct {
	ID                *int       `db:"id" json:"id"`
	Type              *string    `db:"type" json:"type"`
	Total             *float64   `db:"total" json:"total"`
	Direction         *string    `db:"direction" json:"direction"`
	CounterpartyType  *string    `db:"counterparty_type" json:"counterparty_type"`
	CounterpartyName  *string    `db:"counterparty_name" json:"counterparty_name"`
	CounterpartyImg   *string    `db:"counterparty_img" json:"counterparty_img,omitempty"`
	CounterpartyPhone *string    `db:"counterparty_phone" json:"counterparty_phone,omitempty"`
	CreatedAt         *time.Time `db:"created_at" json:"created_at"`
}

type DailySummary struct {
	Date         time.Time `json:"date" example:"2025-09-29"`
	TotalExpense int       `json:"total_expense" example:"75000"`
	TotalIncome  int       `json:"total_income" example:"100000"`
}

type WeeklySummary struct {
	WeekStart    time.Time `json:"week_start" example:"2025-09-23"`
	WeekEnd      time.Time `json:"week_end" example:"2025-09-29"`
	TotalExpense int       `json:"total_expense" example:"250000"`
	TotalIncome  int       `json:"total_income" example:"320000"`
}
