package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prospera/internals/repositories"
	"github.com/prospera/internals/utils"
)

type UserHandler struct {
	ur *repositories.UserRepository
}

func NewUserHandler(ur *repositories.UserRepository) *UserHandler {
	return &UserHandler{ur: ur}
}

func (uh *UserHandler) HandlerGetUsers(ctx *gin.Context) {
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		return
	}

	users, err := uh.ur.GetUser(ctx.Request.Context(), uid)
	if err != nil {
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}
