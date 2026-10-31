package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Meta struct {
	Page         int   `json:"page,omitempty"`
	PerPage      int   `json:"per_page,omitempty"`
	TotalRecords int64 `json:"total_records,omitempty"`
	TotalPages   int   `json:"total_pages,omitempty"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

func JSONSuccess(c *gin.Context, httpCode int, message string, data interface{}, meta *Meta) {
	c.JSON(httpCode, SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func JSONError(c *gin.Context, httpCode int, code, message string, details interface{}) {
	c.JSON(httpCode, ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func InternalServerError(c *gin.Context, message string) {
	JSONError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message, nil)
}

func BadRequest(c *gin.Context, code, message string, details interface{}) {
	JSONError(c, http.StatusBadRequest, code, message, details)
}

func Unauthorized(c *gin.Context, message string) {
	JSONError(c, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

func Forbidden(c *gin.Context, message string) {
	JSONError(c, http.StatusForbidden, "FORBIDDEN", message, nil)
}

func NotFound(c *gin.Context, message string) {
	JSONError(c, http.StatusNotFound, "NOT_FOUND", message, nil)
}
