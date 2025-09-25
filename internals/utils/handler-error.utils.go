package utils

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/prospera/internals/models"
)

func HandleError(ctx *gin.Context, code int, status, err, err_real string) {
	log.Printf("%s ----- Cause: %s\n", status, err_real)
	ctx.JSON(code, models.NewErrorResponse(status, err, code))
}

func HandleMiddlewareError(ctx *gin.Context, code int, status, err string) {
	log.Printf("%s ----- Cause: %s\n", status, err)
	ctx.AbortWithStatusJSON(code, models.NewErrorResponse(status, err, code))
}
