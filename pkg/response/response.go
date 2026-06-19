package response

import (
	"net/http"

	"github.com/dip-roy/go-backend/pkg/apperror"
	"github.com/gin-gonic/gin"
)

type Meta struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Pages int64 `json:"pages"`
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type ErrorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Success: true, Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Success: true, Data: data})
}

func Paginated(c *gin.Context, data interface{}, meta Meta) {
	c.JSON(http.StatusOK, Response{Success: true, Data: data, Meta: &meta})
}

func Err(c *gin.Context, err error) {
	appErr, ok := apperror.As(err)
	if !ok {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   &ErrorBody{Code: "INTERNAL_ERROR", Message: "internal server error"},
		})
		return
	}
	c.JSON(appErr.HTTPStatus, Response{
		Success: false,
		Error:   &ErrorBody{Code: appErr.Code, Message: appErr.Message},
	})
}

func ValidationError(c *gin.Context, details interface{}) {
	c.JSON(http.StatusUnprocessableEntity, Response{
		Success: false,
		Error: &ErrorBody{
			Code:    "VALIDATION_ERROR",
			Message: "validation failed",
			Details: details,
		},
	})
}
