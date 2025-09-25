package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prospera/internals/models"
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
		log.Println("get user token error: ", err)
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable to get user's token")
	}

	users, err := uh.ur.GetUser(ctx.Request.Context(), uid)
	if err != nil {
		log.Println("get user error: ", err)
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable get users")
		return
	}

	ctx.JSON(http.StatusOK, models.Response[[]models.User]{
		Success: true,
		Message: "User's list",
		Data:    users,
	})
}

func (uh *UserHandler) HandleGetUserTransactionsHistory(ctx *gin.Context) {
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		log.Println("", err.Error())
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable get user's history transaction")
		return
	}

	history, err := uh.ur.GetUserHistoryTransactions(ctx, uid, 0, 0)
	if err != nil {
		log.Println("get user's history transaction error: ", err.Error())
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable get user's history transaction")
		return
	}

	ctx.JSON(http.StatusOK, models.Response[models.UserHistoryTransactions]{
		Success: true,
		Message: "success",
		Data:    history,
	})
}
