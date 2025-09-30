package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prospera/internals/models"
	"github.com/prospera/internals/pkg"
	"github.com/prospera/internals/repositories"
	"github.com/prospera/internals/utils"
	"github.com/redis/go-redis/v9"
)

type AuthHandler struct {
	Repo *repositories.Auth
	rdb  *redis.Client
}

func NewAuthHandler(repo *repositories.Auth, rdb *redis.Client) *AuthHandler {
	return &AuthHandler{Repo: repo, rdb: rdb}
}

// Register godoc
//
//	@Summary		Register new user
//	@Description	Create a new user account by providing email and password
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.RegisterRequest	true	"User registration payload"
//	@Success		201		{object}	models.Response			"Register account successful"
//	@Failure		400		{object}	models.Response			"failed binding data"
//	@Failure		500		{object}	models.Response			"failed hashed password or Email is already registered"
//	@Router			/auth/register [post]
func (h *AuthHandler) Register(ctx *gin.Context) {
	var req models.RegisterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "failed binding data", err)
		return
	}

	// Hash Password
	hashConfig := pkg.NewHashConfig()
	hashConfig.UseRecommended()
	hashedPassword, err := hashConfig.GenHash(req.Password)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed hashed password", err)
		return
	}

	listId, err := h.Repo.Register(ctx, req.Email, hashedPassword)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "Email is already registered", err)
		return
	}

	for _, v := range listId {
		var redisKey = fmt.Sprintf("Prospera-AllUser-%d", v)
		if err := utils.InvalidateCache(ctx, h.rdb, redisKey); err != nil {
			log.Println("Failed invalidate cache:", err)
		}
	}

	ctx.JSON(http.StatusCreated, models.Response{
		Success: true,
		Message: "Register account successful",
	})
}

// Login godoc
//
//	@Summary		Login user
//	@Description	Authenticate user and return JWT token if successful
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.LoginRequest		true	"User login payload"
//	@Success		200		{object}	models.ResponseLogin	"Login successful"
//	@Failure		400		{object}	models.Response			"failed binding data"
//	@Failure		401		{object}	models.Response			"invalid username or password"
//	@Failure		500		{object}	models.Response			"user not found or failed to generate token"
//	@Router			/auth [post]
func (h *AuthHandler) Login(ctx *gin.Context) {
	var req models.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "failed binding data", err)
		return
	}

	// Cari akun
	userID, hashedPassword, isPinExist, err := h.Repo.Login(ctx.Request.Context(), req.Email)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "user not found", err)
		return
	}
	if userID == 0 {
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "user not found", err)
		return
	}

	// Verifikasi password
	hashConfig := pkg.NewHashConfig()
	match, err := hashConfig.ComparePasswordAndHash(req.Password, hashedPassword)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed compare password", err)
		return
	}
	if !match {
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "invalid username or password", errors.New("invalid password"))
		return
	}

	// Generate JWT
	claims := pkg.NewJWTClaims(userID, req.Email)
	token, err := claims.GenToken()
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed generate token", err)
		return
	}

	ctx.JSON(http.StatusOK, models.ResponseLogin{
		Success:    true,
		Message:    "Login successful",
		Token:      token,
		IsPinExist: isPinExist,
		Email:      claims.Email,
	})
}

// Logout godoc
//
//	@Summary		Logout user
//	@Description	Invalidate the current JWT by blacklisting the token
//	@Tags			Auth
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	models.Response	"Successfully logged out"
//	@Failure		401	{object}	models.Response	"Unauthorized or token expired"
//	@Failure		500	{object}	models.Response	"Failed to blacklist token"
//	@Router			/auth [delete]
func (h *AuthHandler) Logout(ctx *gin.Context) {
	token, err := utils.GetToken(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "failed get token", err)
		return
	}

	expiresAt, err := utils.GetExpiredFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "failed get expired time token", err)
		return
	}

	expiresIn := time.Until(expiresAt)
	if expiresIn <= 0 {
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "token already expired", err)
		return
	}

	if err = h.Repo.Logout(context.Background(), token, expiresIn); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed to blacklist token", err)
		return
	}

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Successfully logged out",
	})
}

// VerifyPIN godoc
//
//	@Summary		Verify user's PIN
//	@Description	Validate the input PIN against the stored hash for the authenticated user
//	@Tags			Auth
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.PINRequest	true	"PIN verification payload"
//	@Success		200		{object}	models.Response		"Success Verify PIN"
//	@Failure		400		{object}	models.Response		"PIN does not match or invalid request"
//	@Failure		500		{object}	models.Response		"Internal server error"
//	@Router			/auth/verify-pin [post]
func (h *AuthHandler) VerifyPIN(c *gin.Context) {
	var req models.PINRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Bad Request", "invalid request", err)
		return
	}

	// Ambil ID dari token
	id, err := utils.GetUserIDFromJWT(c)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Internal Server Error", "unable to get user's token", err)
		return
	}

	// Ambil pin yang tersimpan dari repo
	storedPIN, err := h.Repo.VerifyUserPIN(c.Request.Context(), id)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Internal Server Error", "failed to fetch account", err)
		return
	}

	// Compare di handler
	hashConfig := pkg.NewHashConfig()
	hashConfig.UseRecommended()
	valid, err := hashConfig.ComparePasswordAndHash(req.PIN, storedPIN)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Internal Server Error", "failed to compare pin", err)
		return
	}

	if !valid {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "PIN does not match",
			Data:    valid,
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Success Verify PIN",
		Data:    valid,
	})
}

// UpdatePIN godoc
//
//	@Summary		Register or update user's PIN
//	@Description	Update or create PIN for authenticated user
//	@Tags			Auth
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.PINRequest	true	"PIN update payload"
//	@Success		201		{object}	models.Response		"Register PIN successful"
//	@Failure		400		{object}	models.Response		"failed binding data"
//	@Failure		500		{object}	models.Response		"failed to update PIN"
//	@Router			/auth/update-pin [post]
func (h *AuthHandler) UpdatePIN(ctx *gin.Context) {
	var req models.PINRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "failed binding data", err)
		return
	}

	//Get ID from Token
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable to get user's token", err)
		return
	}

	// Hash Password
	hashConfig := pkg.NewHashConfig()
	hashConfig.UseRecommended()
	hashedPIN, err := hashConfig.GenHash(req.PIN)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed hashed password", err)
		return
	}

	if err := h.Repo.UpdatePIN(ctx, hashedPIN, uid); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed created account", err)
		return
	}

	ctx.JSON(http.StatusCreated, models.Response{
		Success: true,
		Message: "Register PIN successful",
		Data:    "",
	})
}

// ChangePIN godoc
//
//	@Summary		Change user's PIN
//	@Description	Verify old PIN and update to new PIN for authenticated user
//	@Tags			Auth
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.ChangePINRequest	true	"Change PIN payload"
//	@Success		201		{object}	models.Response			"Change PIN successful"
//	@Failure		400		{object}	models.Response			"PIN does not match or failed binding data"
//	@Failure		500		{object}	models.Response			"Internal server error"
//	@Router			/auth/change-pin [post]
func (h *AuthHandler) ChangePIN(ctx *gin.Context) {
	// Used in profile/change-pin
	// Check old pin, if matches then update to new pin
	var req models.ChangePINRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "failed binding data", err)
		return
	}

	// Get ID from Token
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable to get user's token", err)
		return
	}

	// Verify Old Pin
	// Ambil pin yang tersimpan dari repo
	storedPIN, err := h.Repo.VerifyUserPIN(ctx.Request.Context(), uid)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed to fetch account", err)
		return
	}

	// Compare di handler
	hashConfig := pkg.NewHashConfig()
	hashConfig.UseRecommended()
	valid, err := hashConfig.ComparePasswordAndHash(req.OldPIN, storedPIN)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed to compare pin", err)
		return
	}

	if !valid {
		ctx.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "PIN does not match",
			Data:    valid,
		})
		return
	}

	// Hash Password
	hashConfig.UseRecommended()
	hashedPIN, err := hashConfig.GenHash(req.NewPIN)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed hashed password", err)
		return
	}

	if err := h.Repo.UpdatePIN(ctx, hashedPIN, uid); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed created account", err)
		return
	}

	ctx.JSON(http.StatusCreated, models.Response{
		Success: true,
		Message: "Change PIN successful",
		Data:    "",
	})
}

// func (h *AuthHandler) CheckEmail(ctx *gin.Context) {
// 	email := ctx.Query("email")
// 	if email == "" {
// 		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "missing email query param", nil)
// 		return
// 	}

// 	exists, err := h.Repo.CheckEmail(ctx, email)
// 	if err != nil {
// 		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed to check email", err)
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, models.Response{
// 		Success: true,
// 		Message: "Email check successful",
// 		Data: map[string]bool{
// 			"exists": exists,
// 		},
// 	})
// }

// Forgot godoc
//
//	@Summary		Request reset link for password or PIN
//	@Description	Sends a reset link to the registered email for password or PIN reset
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.ForgotRequest	true	"Forgot password or PIN request"
//	@Success		200		{object}	models.Response			"Reset link sent to your email"
//	@Failure		400		{object}	models.Response			"Invalid request"
//	@Failure		404		{object}	models.Response			"Email not found"
//	@Router			/auth/forgot [post]
func (h *AuthHandler) Forgot(ctx *gin.Context) {
	var req models.ForgotRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.HandleError(ctx, 400, "Bad Request", "Invalid request", err)
		return
	}

	if req.Type != "password" && req.Type != "pin" {
		utils.HandleError(ctx, 400, "Bad Request", "Invalid request", fmt.Errorf("type must be password or pin"))
		return
	}

	// Cek user
	user, err := h.Repo.FindByEmail(req.Email)
	if err != nil {
		utils.HandleError(ctx, 404, "Not Found", "email not found", err)
		return
	}

	// Generate token
	token := utils.GenerateRandomToken()
	h.Repo.SaveResetToken(user.ID, token, time.Now().Add(15*time.Minute))

	// Kirim email
	var resetURL string
	if req.Type == "password" {
		resetURL = fmt.Sprintf("%sauth/create/password?token=%s", os.Getenv("URL_BASE"), token)
	} else {
		resetURL = fmt.Sprintf("%sauth/create/pin?token=%s", os.Getenv("URL_BASE"), token)
	}

	utils.SendResetPasswordEmail(user.Email, resetURL, req.Type)

	ctx.JSON(200, models.Response{
		Success: true,
		Message: "Reset link sent to your email",
	})
}

// ResetPIN godoc
//
//	@Summary		Reset user's PIN using reset token
//	@Description	Reset PIN by providing new PIN and valid reset token
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.PINResetRequest	true	"Reset PIN payload"
//	@Success		201		{object}	models.Response			"Reset PIN successful"
//	@Failure		400		{object}	models.Response			"Failed binding data"
//	@Failure		500		{object}	models.Response			"Failed to reset PIN"
//	@Router			/auth/reset-pin [post]
func (h *AuthHandler) ResetPIN(ctx *gin.Context) {
	var req models.PINResetRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "failed binding data", err)
		return
	}
	// Hash Password
	hashConfig := pkg.NewHashConfig()
	hashConfig.UseRecommended()
	hashedPIN, err := hashConfig.GenHash(req.PIN)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed hashed pin", err)
		return
	}

	if err := h.Repo.ResetPIN(ctx, hashedPIN, req.Token); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed reset pin", err)
		return
	}

	ctx.JSON(http.StatusCreated, models.Response{
		Success: true,
		Message: "Reset PIN successful",
	})
}

// ResetPassword godoc
//
//	@Summary		Reset user's password using reset token
//	@Description	Reset password by providing new password and valid reset token
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.PasswordResetRequest	true	"Reset Password payload"
//	@Success		201		{object}	models.Response				"Reset Password successful"
//	@Failure		400		{object}	models.Response				"Failed binding data"
//	@Failure		500		{object}	models.Response				"Failed to reset password"
//	@Router			/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(ctx *gin.Context) {
	var req models.PasswordResetRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "failed binding data", err)
		return
	}
	// Hash Password
	hashConfig := pkg.NewHashConfig()
	hashConfig.UseRecommended()
	hashedPassword, err := hashConfig.GenHash(req.Password)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed hashed password", err)
		return
	}

	if err := h.Repo.ResetPassword(ctx, hashedPassword, req.Token); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed reset password", err)
		return
	}

	ctx.JSON(http.StatusCreated, models.Response{
		Success: true,
		Message: "Reset Password successful",
	})
}
