package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"goshop/internal/domain/model"
	"goshop/internal/domain/service"
	"goshop/pkg/response"
)

type HomeHandler struct {
	platform *service.PlatformService
	product  *service.ProductService
}

func NewHomeHandler(platform *service.PlatformService, product *service.ProductService) *HomeHandler {
	return &HomeHandler{platform: platform, product: product}
}

func (h *HomeHandler) SiteSettings(c *gin.Context) {
	settings, err := h.platform.GetSiteSettings(c.Request.Context())
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, settings)
}

func (h *HomeHandler) Banners(c *gin.Context) {
	position := c.DefaultQuery("position", model.BannerPositionHome)
	list, err := h.platform.ListBanners(c.Request.Context(), true, position)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *HomeHandler) Categories(c *gin.Context) {
	parentID, _ := strconv.ParseUint(c.DefaultQuery("parent_id", "0"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0"))
	var list interface{}
	var err error
	if parentID > 0 {
		list, err = h.platform.ListCategoriesByParent(c.Request.Context(), parentID, limit)
	} else {
		if limit <= 0 {
			limit = 8
		}
		list, err = h.platform.ListTopCategories(c.Request.Context(), limit)
	}
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *HomeHandler) CategoryTree(c *gin.Context) {
	tree, err := h.platform.ListCategoryTree(c.Request.Context(), true)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, tree)
}

func (h *HomeHandler) FeaturedProducts(c *gin.Context) {
	featured, err := h.platform.ListFeaturedPublic(c.Request.Context())
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	if len(featured) == 0 {
		response.OK(c, []service.ProductListItem{})
		return
	}
	ids := make([]uint64, 0, len(featured))
	sortMap := make(map[uint64]int)
	for _, f := range featured {
		ids = append(ids, f.ProductID)
		sortMap[f.ProductID] = f.Sort
	}
	items, _, err := h.product.ListMarketplace(c.Request.Context(), 1, len(ids), service.MarketplaceQuery{ProductIDs: ids})
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	// preserve featured sort order
	byID := make(map[uint64]service.ProductListItem, len(items))
	for _, it := range items {
		byID[it.ID] = it
	}
	ordered := make([]service.ProductListItem, 0, len(featured))
	for _, f := range featured {
		if it, ok := byID[f.ProductID]; ok {
			ordered = append(ordered, it)
		}
	}
	response.OK(c, ordered)
}
