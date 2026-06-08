package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"goshop/internal/domain/model"
	"goshop/internal/domain/service"
	"goshop/internal/middleware"
	"goshop/pkg/response"
)

type ProductHandler struct {
	svc      *service.ProductService
	platform *service.PlatformService
}

func NewProductHandler(svc *service.ProductService, platform *service.PlatformService) *ProductHandler {
	return &ProductHandler{svc: svc, platform: platform}
}

type skuReq struct {
	Spec          string            `json:"spec"`
	SpecValues    map[string]string `json:"spec_values"`
	Price         int64             `json:"price"`
	OriginalPrice int64             `json:"original_price"`
	Stock         int64             `json:"stock"`
	Image         string            `json:"image"`
}

type specDimensionReq struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type CreateProductReq struct {
	CategoryID  uint64             `json:"category_id"`
	Name        string             `json:"name" binding:"required"`
	Subtitle    string             `json:"subtitle"`
	Description string             `json:"description"`
	CoverImage  string             `json:"cover_image"`
	Images      string             `json:"images"`
	Specs       []specDimensionReq `json:"specs"`
	SKUs        []skuReq           `json:"skus" binding:"required,min=1"`
}

type UpdateProductReq struct {
	CategoryID  uint64             `json:"category_id"`
	Name        string             `json:"name" binding:"required"`
	Subtitle    string             `json:"subtitle"`
	Description string             `json:"description"`
	CoverImage  string             `json:"cover_image"`
	Images      string             `json:"images"`
	Specs       []specDimensionReq `json:"specs"`
	SKUs        []skuReq           `json:"skus" binding:"required,min=1"`
}

type UpdateStatusReq struct {
	Status int8 `json:"status" binding:"required"`
}

func (h *ProductHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	list, total, err := h.svc.ListAdmin(c.Request.Context(), middleware.GetAdminDataTenantID(c), page, pageSize, keyword)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func (h *ProductHandler) Detail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	p, skus, err := h.svc.AdminDetail(c.Request.Context(), middleware.GetAdminDataTenantID(c), id)
	if err != nil {
		response.Fail(c, 40402, "商品不存在")
		return
	}
	response.OK(c, gin.H{"product": p, "skus": skus})
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req CreateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	tenantID := middleware.GetAdminDataTenantID(c)
	p := &model.Product{
		TenantScoped: model.TenantScoped{TenantID: tenantID},
		CategoryID:   req.CategoryID,
		Name:         req.Name,
		Subtitle:     req.Subtitle,
		Description:  req.Description,
		CoverImage:   req.CoverImage,
		Images:       req.Images,
		Specs:        specsToJSON(req.Specs),
		Status:       1,
	}
	skus, err := toSKUs(req.Specs, req.SKUs)
	if err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	if err := h.platform.ValidateCategoryID(c.Request.Context(), req.CategoryID); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	if err := h.svc.Create(c.Request.Context(), p, skus); err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, p)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req UpdateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	tenantID := middleware.GetAdminDataTenantID(c)
	p := &model.Product{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Subtitle:    req.Subtitle,
		Description: req.Description,
		CoverImage:  req.CoverImage,
		Images:      req.Images,
		Specs:       specsToJSON(req.Specs),
	}
	skus, err := toSKUs(req.Specs, req.SKUs)
	if err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	if err := h.platform.ValidateCategoryID(c.Request.Context(), req.CategoryID); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	if err := h.svc.Update(c.Request.Context(), tenantID, id, p, skus); err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *ProductHandler) UpdateStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req UpdateStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	if req.Status != 0 && req.Status != 1 {
		response.Fail(c, 40000, "status 无效")
		return
	}
	if err := h.svc.UpdateStatus(c.Request.Context(), middleware.GetAdminDataTenantID(c), id, req.Status); err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, nil)
}

func specsToJSON(specs []specDimensionReq) string {
	if len(specs) == 0 {
		return ""
	}
	dims := make([]service.SpecDimension, 0, len(specs))
	for _, s := range specs {
		dims = append(dims, service.SpecDimension{Name: s.Name, Values: s.Values})
	}
	return service.MarshalSpecsDimensions(dims)
}

func toSKUs(specs []specDimensionReq, items []skuReq) ([]model.ProductSKU, error) {
	inputs := make([]service.SKUInputItem, 0, len(items))
	for _, s := range items {
		inputs = append(inputs, service.SKUInputItem{
			Spec:          s.Spec,
			SpecValues:    s.SpecValues,
			Price:         s.Price,
			OriginalPrice: s.OriginalPrice,
			Stock:         s.Stock,
			Image:         s.Image,
		})
	}
	dims := make([]service.SpecDimension, 0, len(specs))
	for _, s := range specs {
		dims = append(dims, service.SpecDimension{Name: s.Name, Values: s.Values})
	}
	return service.BuildSKUsFromDimensions(dims, inputs)
}
