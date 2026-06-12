package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/nailx/backend/internal/middleware"
	"github.com/nailx/backend/internal/service"
)

type DesignHandler struct {
	svc service.DesignService
}

func NewDesignHandler(svc service.DesignService) *DesignHandler {
	return &DesignHandler{svc: svc}
}

// POST /api/v1/designs
func (h *DesignHandler) Register(c echo.Context) error {
	var req struct {
		Title           string          `json:"title" validate:"required,max=200"`
		Description     *string         `json:"description"`
		PreviewImageURL string          `json:"preview_image_url" validate:"required,url"`
		DesignData      map[string]any  `json:"design_data" validate:"required"`
		ParentIPID      *string         `json:"parent_ip_id"`
		GenderTag       string          `json:"gender_tag" validate:"required,oneof=feminine masculine neutral"`
		StyleTags       []string        `json:"style_tags"`
		IsPublic        bool            `json:"is_public"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}

	designDataBytes, _ := json.Marshal(req.DesignData)

	designID, jobID, err := h.svc.RegisterDesign(c.Request().Context(), middleware.GetUserID(c), service.RegisterDesignRequest{
		Title:           req.Title,
		Description:     req.Description,
		PreviewImageURL: req.PreviewImageURL,
		DesignData:      designDataBytes,
		ParentIPID:      req.ParentIPID,
		GenderTag:       req.GenderTag,
		StyleTags:       req.StyleTags,
		IsPublic:        req.IsPublic,
	})
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusAccepted, map[string]any{
		"design_ip_id": designID,
		"status":       "pending",
		"job_id":       jobID,
	})
}

// GET /api/v1/designs/similarity-check
func (h *DesignHandler) SimilarityCheck(c echo.Context) error {
	jobID := c.QueryParam("job_id")
	if jobID == "" {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", "job_id is required")
	}
	result, err := h.svc.GetSimilarityCheckStatus(c.Request().Context(), jobID)
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// GET /api/v1/designs
func (h *DesignHandler) ListFeed(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit == 0 {
		limit = 20
	}
	designs, nextCur, err := h.svc.ListFeed(c.Request().Context(), service.DesignFeedFilter{
		GenderTag: strPtr(c.QueryParam("gender_tag")),
		Status:    strPtr(c.QueryParam("status")),
		Sort:      c.QueryParam("sort"),
		Cursor:    c.QueryParam("cursor"),
		Limit:     limit,
	})
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"designs": designs, "next_cursor": nextCur})
}

// GET /api/v1/designs/:id
func (h *DesignHandler) GetDesign(c echo.Context) error {
	design, nodes, err := h.svc.GetDesign(c.Request().Context(), c.Param("id"))
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{
		"design":        design,
		"royalty_nodes": nodes,
	})
}

// GET /api/v1/users/:id/designs
func (h *DesignHandler) ListByCreator(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit == 0 {
		limit = 20
	}
	designs, nextCur, err := h.svc.ListByCreator(c.Request().Context(), c.Param("id"), c.QueryParam("cursor"), limit)
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"designs": designs, "next_cursor": nextCur})
}

// PATCH /api/v1/designs/:id
func (h *DesignHandler) Update(c echo.Context) error {
	var req struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		IsPublic    *bool   `json:"is_public"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	design, err := h.svc.UpdateDesign(c.Request().Context(), middleware.GetUserID(c), middleware.GetUserRole(c), c.Param("id"), service.UpdateDesignRequest{
		Title:       req.Title,
		Description: req.Description,
		IsPublic:    req.IsPublic,
	})
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"design": design})
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
