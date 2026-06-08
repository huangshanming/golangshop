package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"goshop/internal/domain/service"
	"goshop/internal/middleware"
	"goshop/pkg/response"
)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

func (h *ProductHandler) ListPublic(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	categoryID, _ := strconv.ParseUint(c.Query("category_id"), 10, 64)
	tenantID, _ := strconv.ParseUint(c.Query("tenant_id"), 10, 64)

	list, total, err := h.svc.ListMarketplace(c.Request.Context(), page, pageSize, service.MarketplaceQuery{
		Keyword: keyword, CategoryID: categoryID, TenantID: tenantID,
	})
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func (h *ProductHandler) DetailPublic(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	p, skus, err := h.svc.DetailPublic(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 40402, "商品不存在")
		return
	}
	response.OK(c, gin.H{"product": p, "skus": skus})
}

func (h *ProductHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	categoryID, _ := strconv.ParseUint(c.Query("category_id"), 10, 64)
	keyword := c.Query("keyword")

	list, total, err := h.svc.ListApp(c.Request.Context(), middleware.GetTenantID(c), page, pageSize, categoryID, keyword)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func (h *ProductHandler) Detail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	p, skus, err := h.svc.Detail(c.Request.Context(), middleware.GetTenantID(c), id)
	if err != nil {
		response.Fail(c, 40402, "商品不存在")
		return
	}
	if p.Status != 1 {
		response.Fail(c, 40402, "商品不存在")
		return
	}
	response.OK(c, gin.H{"product": p, "skus": skus})
}
