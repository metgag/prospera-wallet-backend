package handlers

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prospera/internals/models"
	"github.com/prospera/internals/pkg"
	"github.com/prospera/internals/repositories"
	"github.com/prospera/internals/utils"
	"github.com/redis/go-redis/v9"
)

type UserHandler struct {
	ur  *repositories.UserRepository
	rdb *redis.Client
}

func NewUserHandler(ur *repositories.UserRepository, rdb *redis.Client) *UserHandler {
	return &UserHandler{ur: ur, rdb: rdb}
}

// GET PROFILE
// GetProfile godoc
//
//	@Summary		Get user profile
//	@Description	Retrieve the profile information of the authenticated user
//	@Tags			User
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	models.Response{data=models.Profile}	"Success Get Profile User"
//	@Failure		500	{object}	models.Response							"Internal Server Error"
//	@Router			/user [get]
func (uh *UserHandler) GetProfile(ctx *gin.Context) {
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable to get user's token", err)
		return
	}

	var cachedData models.Profile
	var redisKey = fmt.Sprintf("Prospera-Profile-%d", uid)
	if err := utils.CacheHit(ctx.Request.Context(), uh.rdb, redisKey, &cachedData); err == nil {
		ctx.JSON(http.StatusOK, models.Response{
			Success: true,
			Message: "Success Get Profile User (from cache)",
			Data:    cachedData,
		})
		return
	}

	profile, err := uh.ur.GetProfile(ctx.Request.Context(), uid)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable get profile user", err)
		// return
	}

	if err := utils.RenewCache(ctx.Request.Context(), uh.rdb, redisKey, profile, 10); err != nil {
		log.Println("Failed to set redis cache:", err)
	}

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Success Get Profile User",
		Data:    *profile,
	})
}

// UPDATE PROFILE
// UpdateProfile godoc
//
//	@Summary		Update user profile
//	@Description	Update user's fullname, phone, and avatar image
//	@Tags			User
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			fullname	formData	string			false	"Full name of the user"
//	@Param			phone		formData	string			false	"Phone number of the user"
//	@Param			img			formData	file			false	"Avatar image file"
//	@Success		200			{object}	models.Response	"Profile updated successfully"
//	@Failure		400			{object}	models.Response	"Bad Request"
//	@Failure		401			{object}	models.Response	"Unauthorized"
//	@Failure		500			{object}	models.Response	"Internal Server Error"
//	@Router			/user/ [patch]
func (uh *UserHandler) UpdateProfile(ctx *gin.Context) {
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "invalid token", err)
		return
	}

	updates := make(map[string]interface{})

	// ambil field dari form-data (jika ada)
	if fullname := ctx.PostForm("fullname"); fullname != "" {
		updates["fullname"] = fullname
	}
	if phone := ctx.PostForm("phone"); phone != "" {
		updates["phone"] = phone
	}

	// upload image jika ada
	file, err := ctx.FormFile("img")
	if err == nil {
		destDir := "public/profile"
		filename := fmt.Sprintf("profile_%d", uid)

		path, saveErr := utils.SaveUploadedFile(ctx, file, destDir, filename)
		if saveErr != nil {
			utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "Upload Failed", saveErr)
			return
		}

		updates["img"] = path
	}

	// update ke DB
	if err := uh.ur.UpdateProfile(ctx.Request.Context(), uid, updates); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed to update profile", err)
		return
	}

	var redisKey = fmt.Sprintf("Prospera-Profile-%d", uid)
	if err := utils.InvalidateCache(ctx, uh.rdb, redisKey); err != nil {
		log.Println("Failed invalidate cache:", err)
	}

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Profile updated successfully",
	})
}

// GET ALL USERS
// GetAllUsers godoc
//
//	@Summary		Get all users
//	@Description	Retrieve a list of all users
//	@Tags			User
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	models.Response{data=[]models.User}	"User's list"
//	@Failure		401	{object}	models.Response						"Unauthorized"
//	@Failure		500	{object}	models.Response						"Internal Server Error"
//	@Router			/user/all [get]
func (uh *UserHandler) GetAllUsers(ctx *gin.Context) {
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable to get user's token", err)
		return
	}

	// var cachedData []models.User
	// var redisKey = fmt.Sprintf("Prospera-AllUser-%d", uid)
	// if err := utils.CacheHit(ctx.Request.Context(), uh.rdb, redisKey, &cachedData); err == nil {
	// 	ctx.JSON(http.StatusOK, models.Response[[]models.User]{
	// 		Success: true,
	// 		Message: "Success Get All Users (from cache)",
	// 		Data:    cachedData,
	// 	})
	// 	return
	// }

	users, err := uh.ur.GetAllUser(ctx.Request.Context(), uid)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable get users", err)
		return
	}

	// if err := utils.RenewCache(ctx.Request.Context(), uh.rdb, redisKey, users, 10); err != nil {
	// 	log.Println("Failed to set redis cache:", err)
	// }

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "User's list",
		Data:    users,
	})
}

// GET HISTORY TRANSACTIONS
// GetUserHistoryTransactions godoc
//
//	@Summary		Get user transaction history
//	@Description	Returns paginated list of user's transaction history
//	@Tags			User
//	@Security		BearerAuth
//	@Produce		json
//	@Param			page	query		int	false	"Page number"
//	@Success		200		{object}	models.PaginatedResponse{data=[]models.TransactionHistory}
//	@Failure		400		{object}	models.Response
//	@Failure		500		{object}	models.Response
//	@Router			/user/history [get]
func (h *UserHandler) GetUserHistoryTransactions(c *gin.Context) {
	// Ambil user_id dari Token
	userID, err := utils.GetUserIDFromJWT(c)
	if err != nil {
		utils.HandleError(c, http.StatusUnauthorized, "Unauthorized", "user not found", err)
		return
	}

	// Ambil query param page, default = 1
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit := 7
	offset := (page - 1) * limit

	if page == 1 {
		var cachedData []models.TransactionHistory
		var redisKey = fmt.Sprintf("Prospera-HistoryTransaction-%d-%d", page, userID)
		if err := utils.CacheHit(c.Request.Context(), h.rdb, redisKey, &cachedData); err == nil {
			c.JSON(http.StatusOK, models.Response{
				Success: true,
				Message: "Success Get History Transaction (from cache)",
				Data:    cachedData,
			})
			return
		}
	}

	// Ambil data transaksi + total count
	transactions, total, err := h.ur.GetUserHistoryTransactions(c.Request.Context(), userID, limit, offset)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Internal Server Error", "failed get history transaction", err)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	if page == 1 {
		var redisKey = fmt.Sprintf("Prospera-HistoryTransaction-%d-%d", page, userID)
		if err := utils.RenewCache(c.Request.Context(), h.rdb, redisKey, transactions, 10); err != nil {
			log.Println("Failed to set redis cache:", err)
		}
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Success:    true,
		Message:    "Success Load History",
		Data:       transactions,
		TotalPages: totalPages,
		Page:       page,
	})
}

// DELETE HISTORY TRANSACTIONS
// HandleSoftDeleteTransaction godoc
//
//	@Summary		Soft delete a transaction from user history
//	@Description	Marks a transaction as deleted for the authenticated user (soft delete)
//	@Tags			User
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Transaction ID"
//	@Produce		json
//	@Success		200	{object}	models.Response	"Success response"
//	@Failure		400	{object}	models.Response	"Bad Request, invalid transaction id"
//	@Failure		401	{object}	models.Response	"Unauthorized, invalid or missing token"
//	@Failure		500	{object}	models.Response	"Internal Server Error"
//	@Router			/user/history/{id} [delete]
func (uh *UserHandler) HandleSoftDeleteTransaction(ctx *gin.Context) {
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable to get user's token", err)
		return
	}

	transId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "invalid param input", err)
		return
	}

	if err := uh.ur.SoftDeleteTransaction(ctx.Request.Context(), uid, transId); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable to delete history", err)
	}

	var redisKey = fmt.Sprintf("Prospera-HistoryTransaction-1-%d", uid)
	if err := utils.InvalidateCache(ctx, uh.rdb, redisKey); err != nil {
		log.Println("Failed invalidate cache:", err)
	}

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "success",
		Data:    "history deleted",
	})
}

// POST CHANGE PASSWORD
// ChangePassword godoc
//
//	@Summary		Change user password
//	@Description	Allows authenticated user to change their password by providing the old password and new password
//	@Tags			User
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		models.ChangePassword	true	"Password change payload"
//	@Success		201		{object}	models.Response			"Password changed successfully"
//	@Failure		400		{object}	models.Response			"Bad request, validation error or old password mismatch"
//	@Failure		401		{object}	models.Response			"Unauthorized, invalid or missing token"
//	@Failure		500		{object}	models.Response			"Internal server error"
//	@Router			/user/password [post]
func (uh *UserHandler) ChangePassword(ctx *gin.Context) {
	var req models.ChangePassword

	if err := ctx.ShouldBind(&req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "failed to bind", err)
		return
	}

	// Get ID from token
	userID, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "user not found", err)
		return
	}

	// Fetch current password hash from DB
	oldPassDB, err := uh.ur.GetPasswordFromID(ctx, userID)
	if err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "bad request", "failed to fetch user credentials", err)
		return
	}

	// Compare old password
	hashCfg := pkg.NewHashConfig()
	isMatched, err := hashCfg.ComparePasswordAndHash(req.OldPassword, oldPassDB)
	if err != nil || !isMatched {
		var handleErr error = err
		if handleErr == nil && !isMatched {
			handleErr = errors.New("old password does not match")
		}
		utils.HandleError(ctx, http.StatusBadRequest, "bad request", "old password does not match", handleErr)
		return
	}

	// Hash new password
	hashCfg.UseRecommended()
	hashedPassword, err := hashCfg.GenHash(req.NewPassword)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "internal server error", "failed to hash new password", err)
		return
	}

	// Update password
	if err := uh.ur.ChangePassword(ctx, userID, hashedPassword); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "internal server error", "failed to change password", err)
		return
	}

	ctx.JSON(http.StatusCreated, models.Response{
		Success: true,
		Message: "Change password successful",
	})
}

// GET SUMMARY
// GetSummary godoc
//
//	@Summary		Get user's financial summary
//	@Description	Get income and expense summary (daily or weekly)
//	@Tags			User
//	@Security		BearerAuth
//	@Produce		json
//	@Param			range	query		string										false	"daily or weekly"	Enums(daily, weekly)
//	@Success		200		{object}	models.Response{data=models.DailySummary}	//		jika				daily
//	@Success		200		{object}	models.Response{data=models.WeeklySummary}	//		jika				weekly
//	@Failure		400		{object}	models.Response
//	@Failure		401		{object}	models.Response
//	@Failure		500		{object}	models.Response
//	@Router			/user/summary [get]
func (h *UserHandler) GetSummary(c *gin.Context) {
	// Ambil user ID
	userID, err := utils.GetUserIDFromJWT(c)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Internal Server Error", "unable to get user's token", err)
		return
	}

	// Ambil query parameter "range", default daily
	rangeType := c.DefaultQuery("range", "daily")

	var cachedData models.Profile
	var redisKey = fmt.Sprintf("Prospera-Summary-%s-%d", rangeType, userID)
	if err := utils.CacheHit(c.Request.Context(), h.rdb, redisKey, &cachedData); err == nil {
		c.JSON(http.StatusOK, models.Response{
			Success: true,
			Message: fmt.Sprintf("Success Get Summary %s (from cache)", rangeType),
			Data:    cachedData,
		})
		return
	}

	switch rangeType {
	case "daily":
		summaries, err := h.ur.GetDailySummary(c, userID)
		if err != nil {
			utils.HandleError(c, http.StatusInternalServerError, "Internal Server Error", "Failed Get Summary Daily", err)
			return
		}
		if err := utils.RenewCache(c.Request.Context(), h.rdb, redisKey, summaries, 10); err != nil {
			log.Println("Failed to set redis cache:", err)
		}
		c.JSON(http.StatusOK, models.Response{
			Success: true,
			Message: "Success Get Summary Daily",
			Data:    summaries,
		})

	case "weekly":
		summaries, err := h.ur.GetWeeklySummary(c, userID)
		if err != nil {
			utils.HandleError(c, http.StatusInternalServerError, "Internal Server Error", "Failed Get Summary Weekly", err)
			return
		}
		if err := utils.RenewCache(c.Request.Context(), h.rdb, redisKey, summaries, 10); err != nil {
			log.Println("Failed to set redis cache:", err)
		}
		c.JSON(http.StatusOK, models.Response{
			Success: true,
			Message: "Success Get Summary Daily",
			Data:    summaries,
		})

	default:
		utils.HandleError(c, http.StatusInternalServerError, "Internal Server Error", "invalid range type, must be 'daily' or 'weekly", err)
		return
	}

}

// GET BALANCE
// GetBalance godoc
//
//	@Summary		Get wallet balance
//	@Description	Fetch the current wallet balance for the authenticated user.
//	@Tags			User
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	models.Response{data=int}	"Success"
//	@Failure		401	{object}	models.Response				"Unauthorized"
//	@Failure		404	{object}	models.Response				"Wallet not found"
//	@Failure		500	{object}	models.Response				"Internal server error"
//	@Router			/user/wallet [get]
func (h *UserHandler) GetBalance(ctx *gin.Context) {
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "invalid token", err)
		return
	}

	var cachedData int
	var redisKey = fmt.Sprintf("Prospera-Balance-%d", uid)
	if err := utils.CacheHit(ctx.Request.Context(), h.rdb, redisKey, &cachedData); err == nil {
		ctx.JSON(http.StatusOK, models.Response{
			Success: true,
			Message: "Success Get Balance (from cache)",
			Data:    cachedData,
		})
		return
	}

	balance, err := h.ur.GetBalanceByWalletID(ctx.Request.Context(), uid)
	if err != nil {
		utils.HandleError(ctx, http.StatusNotFound, "Not Found", "wallet not found", err)
		return
	}

	if err := utils.RenewCache(ctx.Request.Context(), h.rdb, redisKey, balance, 10); err != nil {
		log.Println("Failed to set redis cache:", err)
	}

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Wallet balance fetched successfully",
		Data:    balance,
	})
}

// RemoveAvatar godoc
//
//	@Summary		Remove user avatar
//	@Description	Removes the current avatar of the authenticated user.
//	@Tags			User
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	models.Response{data=string}	"User's avatar removed"
//	@Failure		401	{object}	models.Response					"Unauthorized"
//	@Failure		500	{object}	models.Response					"Internal Server Error"
//	@Router			/user/avatar [delete]
func (h *UserHandler) RemoveAvatar(ctx *gin.Context) {
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "invalid token", err)
		return
	}

	if err := h.ur.DeleteAvatar(ctx.Request.Context(), uid); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "Unable to remove user's avatar", err)
		return
	}

	var redisKey = fmt.Sprintf("Prospera-Profile-%d", uid)
	if err := utils.InvalidateCache(ctx, h.rdb, redisKey); err != nil {
		log.Println("Failed invalidate cache:", err)
	}

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "User's avatar removed",
		Data:    "success",
	})
}

// GET USER BASED FROM ID
// GetUserWithId godoc
//
//	@Summary		Get User Profile By ID
//	@Description	Retrieve a single user by their ID
//	@Tags			User
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	models.Response{data=models.User}
//	@Failure		400	{object}	models.Response
//	@Failure		500	{object}	models.Response
//	@Router			/user/{id} [get]
func (h *UserHandler) GetUserWithId(ctx *gin.Context) {
	uid, _ := strconv.Atoi(ctx.Param("id"))

	user, err := h.ur.GetUserById(ctx.Request.Context(), uid)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "Unable to get user's", err)
		return
	}

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "success",
		Data:    user,
	})
}
