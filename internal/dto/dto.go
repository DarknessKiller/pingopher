package dto

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ToModel[M any] interface {
	ToModel() *M
}

type ValidationError struct {
	MissingFields []string `json:"missing_fields"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("missing required fields: %s", strings.Join(e.MissingFields, ", "))
}

func BindAndMap[D ToModel[M], M any](ctx *gin.Context) (*M, error) {
	var dto D

	if err := ctx.ShouldBindJSON(&dto); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			missingFields := make([]string, 0, len(ve))
			for _, fe := range ve {
				fieldName := fe.Field()
				if fe.Tag() == "required" {
					missingFields = append(missingFields, fieldName)
				}
			}
			if len(missingFields) > 0 {
				return nil, ValidationError{MissingFields: missingFields}
			}
		}
		return nil, err
	}

	return dto.ToModel(), nil
}
