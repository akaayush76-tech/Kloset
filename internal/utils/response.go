package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse represents standardized API response
type APIResponse struct {
	Success    bool            `json:"success"`
	Message    string          `json:"message"`
	Data       interface{}     `json:"data,omitempty"`
	Error      string          `json:"error,omitempty"`
	Pagination *PaginationInfo `json:"pagination,omitempty"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Page        int64 `json:"page"`
	Limit       int64 `json:"limit"`
	Total       int64 `json:"total"`
	Pages       int64 `json:"pages"`
	HasNext     bool  `json:"hasNext"`
	HasPrevious bool  `json:"hasPrev"`
}

// ErrorResponse sends error response
func ErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
		Error:   errorMsg,
	})
}

// SuccessResponse sends success response
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessResponseWithPagination sends success response with pagination
func SuccessResponseWithPagination(c *gin.Context, statusCode int, message string, data interface{}, pagination *PaginationInfo) {
	c.JSON(statusCode, APIResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	})
}

// PaginationMeta calculates pagination metadata
func PaginationMeta(page, limit, total int64) *PaginationInfo {
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 10
	}
	if page < 1 {
		page = 1
	}

	pages := (total + limit - 1) / limit // ceiling division
	if pages < 1 && total > 0 {
		pages = 1
	}

	hasNext := (page * limit) < total
	hasPrev := page > 1

	return &PaginationInfo{
		Page:        page,
		Limit:       limit,
		Total:       total,
		Pages:       pages,
		HasNext:     hasNext,
		HasPrevious: hasPrev,
	}
}

// HTTPErrorHandler handles HTTP errors
func HTTPErrorHandler(c *gin.Context, statusCode int, message string, err error) {
	logger := GetLogger()

	switch statusCode {
	case http.StatusBadRequest:
		logger.Error("Bad Request: %v", err)
	case http.StatusUnauthorized:
		logger.Error("Unauthorized: %v", err)
	case http.StatusForbidden:
		logger.Error("Forbidden: %v", err)
	case http.StatusNotFound:
		logger.Debug("Not Found: %v", err)
	case http.StatusConflict:
		logger.Error("Conflict: %v", err)
	case http.StatusTooManyRequests:
		logger.Debug("Too Many Requests")
	case http.StatusInternalServerError:
		logger.Error("Internal Server Error: %v", err)
	default:
		logger.Error("HTTP %d Error: %v", statusCode, err)
	}

	ErrorResponse(c, statusCode, message, err)
}
