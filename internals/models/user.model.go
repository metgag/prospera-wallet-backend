package models

type User struct {
	FullName string `json:"full_name"`
}

type UserHistoryTransactions struct {
	ID           int           `json:"id"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	ReceiverID      int    `json:"receiver_id"`
	Avatar          string `json:"avatar"`
	FullName        string `json:"receiver_name"`
	PhoneNumber     string `json:"receiver_phone"`
	TransactionType string `json:"transaction_type"`
	Total           int    `json:"total"`
}
