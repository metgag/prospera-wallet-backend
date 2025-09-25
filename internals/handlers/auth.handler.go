package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prospera/internals/models"
	"github.com/prospera/internals/pkg"
	"github.com/prospera/internals/repositories"
	"github.com/prospera/internals/utils"
)

type AuthHandler struct {
	Repo *repositories.Auth
}

func NewAuthHandler(repo *repositories.Auth) *AuthHandler {
	return &AuthHandler{Repo: repo}
}

func (h *AuthHandler) Register(ctx *gin.Context) {
	var req models.RegisterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "failed binding data")
		return
	}

	// Hash Password
	hashConfig := pkg.NewHashConfig()
	hashConfig.UseRecommended()
	hashedPassword, err := hashConfig.GenHash(req.Password)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed hashed password")
		return
	}

	if err := h.Repo.Register(ctx, req.Email, hashedPassword); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, models.Response[string]{
		Success: true,
		Message: "Register successful",
		Data:    "",
	})
}
