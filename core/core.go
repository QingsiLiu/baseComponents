package core

import (
	"net/http"

	"github.com/QingsiLiu/baseComponents/errors"
	"github.com/QingsiLiu/baseComponents/log"
	"github.com/gin-gonic/gin"
)

// Response defines the unified return messages.
// swagger:model
type Response struct {
	// Code defines the business error code.
	Code int `json:"code"`

	// Message contains the detail of this message.
	// This message is suitable to be exposed to external.
	Message string `json:"message"`

	// Data contains the response payload when no error occurred.
	Data interface{} `json:"data"`
}

// WriteResponse write an error or the response data into http response body.
// It use errors.ParseCoder to parse any error into errors.Coder
// errors.Coder contains error code, user-safe error message and http status code.
func WriteResponse(c *gin.Context, err error, data interface{}) {
	if err != nil {
		log.Errorf("%#+v", err)
		coder := errors.ParseCoder(err)
		c.JSON(coder.HTTPStatus(), Response{
			Code:    coder.Code(),
			Message: coder.String(),
			Data:    nil,
		})

		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}
