package handlers

import "github.com/prospera/internals/repositories"

type UserHandler struct {
	ur *repositories.UserRepository
}

func NewUserHandler(ur *repositories.UserRepository) *UserHandler {
	return &UserHandler{ur: ur}
}
