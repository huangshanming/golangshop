package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"goshop/pkg/errcode"
)

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: errcode.OK, Message: "success", Data: data})
}

func Fail(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Body{Code: code, Message: message})
}

func FailWithHTTP(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Body{Code: code, Message: message})
}

func Page(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	OK(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
