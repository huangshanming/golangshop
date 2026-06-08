package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"goshop/internal/domain/service"
	"goshop/internal/middleware"
	"goshop/pkg/response"
)

type PlatformHandler struct {
	platform *service.PlatformService
	product  *service.ProductService
}

func NewPlatformHandler(platform *service.PlatformService, product *service.ProductService) *PlatformHandler {
	return &PlatformHandler{platform: platform, product: product}
}

func (h *PlatformHandler) ListCategories(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	tree, err := h.platform.ListCategoryTree(c.Request.Context(), false)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, tree)
}

func (h *PlatformHandler) CreateCategory(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	var req service.PlatformCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	item, err := h.platform.CreateCategory(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, item)
}

func (h *PlatformHandler) UpdateCategory(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.PlatformCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	if err := h.platform.UpdateCategory(c.Request.Context(), id, req); err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *PlatformHandler) DeleteCategory(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.platform.DeleteCategory(c.Request.Context(), id); err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *PlatformHandler) ListBanners(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	position := c.Query("position")
	list, err := h.platform.ListBanners(c.Request.Context(), false, position)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *PlatformHandler) CreateBanner(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	var req service.BannerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	item, err := h.platform.CreateBanner(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, item)
}

func (h *PlatformHandler) UpdateBanner(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.BannerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	if err := h.platform.UpdateBanner(c.Request.Context(), id, req); err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *PlatformHandler) DeleteBanner(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.platform.DeleteBanner(c.Request.Context(), id); err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *PlatformHandler) GetSiteSettings(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	settings, err := h.platform.GetSiteSettings(c.Request.Context())
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, settings)
}

func (h *PlatformHandler) UpdateSiteSettings(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	var req service.SiteSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	if err := h.platform.UpdateSiteSettings(c.Request.Context(), req); err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *PlatformHandler) ListFeatured(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	list, err := h.platform.ListFeaturedAdmin(c.Request.Context())
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *PlatformHandler) AddFeatured(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	var req service.AddFeaturedReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	item, err := h.platform.AddFeatured(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, item)
}

func (h *PlatformHandler) DeleteFeatured(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.platform.DeleteFeatured(c.Request.Context(), id); err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *PlatformHandler) UpdateFeaturedSort(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	var req struct {
		Items []service.FeaturedSortItem `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	if err := h.platform.UpdateFeaturedSort(c.Request.Context(), req.Items); err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *PlatformHandler) SearchProducts(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	keyword := c.Query("keyword")
	list, err := h.product.SearchGlobal(c.Request.Context(), keyword, 30)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, list)
}
