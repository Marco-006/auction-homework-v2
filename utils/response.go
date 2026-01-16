package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code uint   `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

func Fail(c *gin.Context, errMsg string) {
	c.JSON(http.StatusOK, Response{
		Code: 10000,
		Msg:  errMsg,
		Data: nil,
	})
}
