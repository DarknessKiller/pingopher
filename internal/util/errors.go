package util

import (
	"errors"
	"io"
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
		returnError(http.StatusInternalServerError, err.Error())
	}
}
