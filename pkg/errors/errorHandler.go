package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler(c *gin.Context, err error) {
	var status int
	errorType := GetType(err)
	switch errorType {
	case BadRequest:
		status = http.StatusBadRequest
	case NotFound:
		status = http.StatusNotFound
	case NilUserId:
		status = http.StatusBadRequest
	case NilRole:
		status = http.StatusBadRequest
	case NotSaved:
		status = http.StatusInternalServerError
	case WrongType:
		status = http.StatusBadRequest
	case EmptyData:
		status = http.StatusBadRequest
	default:
		status = http.StatusInternalServerError
	}
	c.Writer.WriteHeader(status)

	response := gin.H{"error": err.Error(), "message": errorType.Message()}

	errorContext := GetErrorContext(err)
	if errorContext != nil {
		response["context"] = errorContext
	}
	c.JSON(status, response)
}
