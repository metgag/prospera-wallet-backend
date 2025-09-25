package handlers

import (
	"github.com/prospera/internals/repositories"
)

type TransactionHandler struct {
	tr *repositories.TransactionRepository
}

func NewTransactionHandler(tr *repositories.TransactionRepository) *TransactionHandler {
	return &TransactionHandler{tr: tr}
}
