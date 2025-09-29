package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prospera/internals/handlers"
	"github.com/prospera/internals/middlewares"
	"github.com/redis/go-redis/v9"

	docs "github.com/prospera/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitRouter(db *pgxpool.Pool, rdb *redis.Client) *gin.Engine {
	router := gin.Default()
	router.Use(middlewares.CORSMiddleware)
	middlewares.InitRedis(rdb)

	docs.SwaggerInfo.BasePath = "/"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	router.Static("profile", "public/profile")

	router.GET("/ws", handlers.WebSocketHandler)

	// Init Route Authentication
	InitAuthRoutes(router, db, rdb)
	InitUserRouter(router, db, rdb)
	InitTransactions(router, db, rdb)
	InitInternalAccountRoute(router, db, rdb)

	return router
}
