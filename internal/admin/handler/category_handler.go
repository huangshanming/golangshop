package handler

import (
	"github.com/gin-gonic/gin"

	"goshop/internal/domain/service"
	"goshop/pkg/response"
)

type CategoryHandler struct {
	svc *service.CategoryService
}

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) List(c *gin.Context) {
	tree, err := h.svc.ListTree(c.Request.Context())
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, tree)
}
