package handler

import (
	"net/http"
	"net/url"
	"shortlink-go/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.IService
}

func NewHandler(srv service.IService) *Handler {
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

// HealthCheck shows the status of the service
// @Summary Show service health status
// @Description Get the health status of the service
// @Tags health
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *Handler) HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// CreateShortLink creates a new short link
// @Summary Create a new short link
// @Description Create a new short link from a given long URL
// @Tags links
// @Accept  json
// @Produce  json
// @Param   request  body      CreateLinkRequest true  "Create Link Request"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /create [post]
func (h *Handler) CreateShortLink(ctx *gin.Context) {
	var request CreateLinkRequest
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

// RedirectToLongURL redirects to the original URL based on the short link provided
// @Summary Redirect to the original URL
// @Description Redirects the request to the original long URL based on the provided short link
// @Tags links
// @Accept  json
// @Produce  json
// @Param   shortLink  path      string  true  "Short Link"
// @Success 307 {header} string Location "Location header with the original URL"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /{shortLink} [get]
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

// GetStats retrieves statistics for a short link
// @Summary Get short link statistics
// @Description Get the statistics of a short link, including its original URL and access count
// @Tags stats
// @Accept  json
// @Produce  json
// @Param   shortLink  path      string  true  "Short Link"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /stats/{shortLink} [get]
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
	res := GetStatsResponse{
		LongURL:     stats.LongURL,
		ShortLink:   shortLink,
		AccessCount: stats.AccessCount,
	}
	ctx.JSON(http.StatusOK, res)
}
