package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func InitRouter(db *pgxpool.Pool, rdb *redis.Client) *gin.Engine {
	router := gin.Default()

	// Init Route Authentication
	InitAuthRoutes(router, db)

	return router
}
