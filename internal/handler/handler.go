package handler

import (
	"net/http"
	"net/url"
	"shortlink-go/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func NewHandler(srv *service.Service) *Handler {
	return &Handler{
		service: srv,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", h.HealthCheck)
	r.POST("/create", h.CreateShortLink)
	r.GET("/:shortLink", h.RedirectToLongURL)
	r.GET("/stats/:shortLink", h.GetStats)
}

func (h *Handler) HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) CreateShortLink(ctx *gin.Context) {
	var request struct {
		LongURL string `json:"long_url"`
	}
	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	// Validate the URL
	if _, err := url.ParseRequestURI(request.LongURL); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	shortLink, err := h.service.CreateShortLink(ctx, request.LongURL)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create short link"})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"shortLink": shortLink})
}

func (h *Handler) RedirectToLongURL(ctx *gin.Context) {
	shortLink := ctx.Param("shortLink")
	longURL, err := h.service.GetLongURL(ctx, shortLink)
	if err != nil {
		if err == service.ErrShortLinkNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Short link not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to redirect to long URL"})
		}
		return
	}
	ctx.Redirect(http.StatusTemporaryRedirect, longURL)
}

func (h *Handler) GetStats(ctx *gin.Context) {
	shortLink := ctx.Param("shortLink")
	stats, err := h.service.GetLinkStats(ctx, shortLink)
	if err != nil {
		if err == service.ErrShortLinkNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Short link not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get link stats"})
		}
		return
	}
	ctx.JSON(http.StatusOK, stats)
}
