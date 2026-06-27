package util

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/DarknessKiller/pingopher/internal/dto"
	"github.com/gin-gonic/gin"
)

func HandleError(ctx *gin.Context, err error) {
	returnError := func(status int, msg string) { ctx.JSON(status, gin.H{"status": "error", "message": msg}) }

	switch {
	case errors.As(err, &dto.ValidationError{}):
		returnError(http.StatusBadRequest, err.Error())
	case errors.Is(err, io.EOF):
	default:
		log.Printf("Internal error: %v", err)
		returnError(http.StatusInternalServerError, "internal server error")
	}
}
