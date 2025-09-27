package utils

import (
	"github.com/go-chi/render"
	"github.com/gin-gonic/gin"
	"net/http"
)

// BizCode 业务码
type BizCode struct {
	Code    int32
	Message string
}

// 业务错误码
var (
	/*通用错误码*/
	Success   = BizCode{Code: 0000, Message: "success"}        // 成功
	Failed    = BizCode{Code: 1000, Message: "failed"}         // 失败
	ParamErr  = BizCode{Code: 1001, Message: "param error"}    // 参数错误
	ServerErr = BizCode{Code: 1002, Message: "server error"}   // 服务器错误
	DBErr     = BizCode{Code: 1003, Message: "database error"} // 数据库错误

	/*业务错误码*/
)

type Resp struct {
	Message string      `json:"msg"`
	Success bool        `json:"success"`
	Code    int32       `json:"code"`
	Data    interface{} `json:"data"`
}

// ErrResp 发生错误响应
func ErrResp(w http.ResponseWriter, r *http.Request, bizCode BizCode, code int) {
	resp := &Resp{
		Message: bizCode.Message,
		Code:    bizCode.Code,
		Success: false,
		Data:    nil,
	}
	render.Status(r, code)
	render.JSON(w, r, resp)
}

// OKResp 成功执行响应
func OKResp(w http.ResponseWriter, r *http.Request, data interface{}) {
	resp := &Resp{
		Message: "success",
		Code:    Success.Code,
		Success: true,
		Data:    data,
	}
	render.JSON(w, r, resp)
}

// GinResp gin框架响应
func GinOKResp(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Resp{
		Message: "success",
		Code:    Success.Code,
		Success: true,
		Data:    data,
	})
}

// GinErrResp gin框架错误响应
func GinErrResp(c *gin.Context, bizCode BizCode, code int) {
	c.JSON(code, Resp{
		Message: bizCode.Message,
		Code:    bizCode.Code,
		Success: false,
		Data:    nil,
	})
}
