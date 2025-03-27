package response

import (
	"net/http"
	"time"

	"go-template/pkg/errcode"

	"github.com/gin-gonic/gin"
)

// Response defines the unified API response structure
type Response struct {
	Code      string      `json:"code"`                 // code identifier
	Message   string      `json:"message"`              // code message
	Data      interface{} `json:"data,omitempty"`       // Response data
	Timestamp int64       `json:"timestamp"`            // Unix timestamp in milliseconds
	RequestID string      `json:"request_id,omitempty"` // Unique request identifier
}

// getRequestID retrieves the request ID from context
func getRequestID(c *gin.Context) string {
	requestID, exists := c.Get("RequestID")
	if !exists {
		return ""
	}
	return requestID.(string)
}

// Ok sends a successful response with data
func Ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      errcode.Ok,
		Message:   errcode.GetMessage(errcode.Ok),
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
		RequestID: getRequestID(c),
	})
}

// OkWithMessage sends a successful response with a custom message and data
func OkWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      errcode.Ok,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
		RequestID: getRequestID(c),
	})
}

// Err sends an error response with a specified code
// Optional custom message can be provided to override the default message
func Err(c *gin.Context, code string, customMsg ...string) {
	message := errcode.GetMessage(code)
	if len(customMsg) > 0 && customMsg[0] != "" {
		message = customMsg[0]
	}

	c.JSON(http.StatusInternalServerError, Response{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UnixMilli(),
		RequestID: getRequestID(c),
	})
}

// ErrWithData sends an error response with additional data
// Optional custom message can be provided to override the default message
func ErrWithData(c *gin.Context, code string, data interface{}, customMsg ...string) {
	message := errcode.GetMessage(code)
	if len(customMsg) > 0 && customMsg[0] != "" {
		message = customMsg[0]
	}

	c.JSON(http.StatusInternalServerError, Response{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
		RequestID: getRequestID(c),
	})
}

// RespondWithError handles various error types and sends appropriate responses
func RespondWithError(c *gin.Context, err error) {
	// Check if it's our custom error type
	if e, ok := err.(*errcode.Error); ok {
		c.JSON(http.StatusInternalServerError, Response{
			Code:      e.Code,
			Message:   e.Message,
			Timestamp: time.Now().UnixMilli(),
			RequestID: getRequestID(c),
		})
		return
	}

	// For other error types, use unknown error code
	c.JSON(http.StatusInternalServerError, Response{
		Code:      errcode.Unknown,
		Message:   err.Error(),
		Timestamp: time.Now().UnixMilli(),
		RequestID: getRequestID(c),
	})
}
