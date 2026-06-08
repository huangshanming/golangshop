package handler

import (
	"github.com/gin-gonic/gin"

	"goshop/internal/domain/service"
	"goshop/internal/middleware"
	"goshop/pkg/response"
)

type UploadHandler struct {
	svc *service.UploadService
}

func NewUploadHandler(svc *service.UploadService) *UploadHandler {
	return &UploadHandler{svc: svc}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	if h.svc == nil {
		response.Fail(c, 50000, "文件存储未配置")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, 40000, "请选择文件")
		return
	}
	f, err := file.Open()
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	defer f.Close()

	result, err := h.svc.UploadImage(c.Request.Context(), middleware.GetAdminDataTenantID(c), file.Filename, f, file.Size)
	if err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	response.OK(c, result)
}
