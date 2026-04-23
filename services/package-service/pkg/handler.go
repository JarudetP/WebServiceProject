package pkg

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GET /api/packages
func (h *Handler) ListPackages(c *gin.Context) {
	packages, err := h.service.ListPackages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, packages)
}

// GET /api/packages/:id
func (h *Handler) GetPackage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid package id"})
		return
	}
	pkg, err := h.service.GetPackage(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pkg)
}

// POST /api/packages/purchase
func (h *Handler) Purchase(c *gin.Context) {
	var req PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Try to get userID from JWT context first
	userID := 0
	if val, exists := c.Get("jwt_user_id"); exists {
		userID = val.(int)
	} else {
		// Fallback to query param for dev/testing
		userIDStr := c.Query("user_id")
		userID, _ = strconv.Atoi(userIDStr)
	}

	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	sub, err := h.service.Purchase(userID, req.PackageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "purchase successful", "subscription": sub})
}

// GET /api/packages/subscription
func (h *Handler) GetActiveSubscription(c *gin.Context) {
	userID := 0
	if val, exists := c.Get("jwt_user_id"); exists {
		userID = val.(int)
	} else {
		userIDStr := c.Query("user_id")
		userID, _ = strconv.Atoi(userIDStr)
	}

	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	sub, err := h.service.GetActiveSubscription(userID)
	if err != nil {
		c.JSON(http.StatusOK, nil)
		return
	}

	c.JSON(http.StatusOK, sub)
}
