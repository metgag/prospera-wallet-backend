package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prospera/internals/handlers"
	"github.com/prospera/internals/repositories"
)

func InitAuthRoutes(router *gin.Engine, db *pgxpool.Pool) {
	repo := repositories.NewAuthRepo(db)
	handler := handlers.NewAuthHandler(repo)

	auth := router.Group("/auth")
	auth.POST("", handler.Login)
	// auth.DELETE("", handler.Logout)
	auth.POST("/register", handler.Register)
}
