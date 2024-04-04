package handler_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"shortlink-go/internal/handler"
	"shortlink-go/internal/model"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) CreateShortLink(ctx context.Context, longURL string) (string, error) {
	args := m.Called(ctx, longURL)
	return args.String(0), args.Error(1)
}

func (m *MockService) GetLongURL(ctx context.Context, shortLink string) (string, error) {
	args := m.Called(ctx, shortLink)
	return args.String(0), args.Error(1)
}

func (m *MockService) GetLinkStats(ctx context.Context, shortLink string) (*model.URL, error) {
	args := m.Called(ctx, shortLink)
	if args.Get(0) != nil {
		return args.Get(0).(*model.URL), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestHandler_CreateShortLink(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create an instance of our test object
	mockService := new(MockService)
	handler := handler.NewHandler(mockService)

	// Mock Service's response
	mockShortLink := "abcd1234"
	mockService.On("CreateShortLink", mock.Anything, "https://example.com").Return(mockShortLink, nil)

	// Create request and recorder
	longURL := `{"long_url":"https://example.com"}`
	req, err := http.NewRequest(http.MethodPost, "/create", bytes.NewBufferString(longURL))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Set up routes
	router.POST("/create", handler.CreateShortLink)

	// Perform request
	router.ServeHTTP(w, req)

	// Assert expectations
	assert.Equal(t, http.StatusCreated, w.Code)
	expectedResponse := `{"shortLink":"abcd1234"}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
	mockService.AssertExpectations(t)
}

func TestHandler_CreateShortLink_InvalidURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := new(MockService)
	h := handler.NewHandler(mockService)

	router.POST("/create", h.CreateShortLink)

	invalidURL := `{"long_url":"not_a_valid_url"}`
	req, _ := http.NewRequest(http.MethodPost, "/create", bytes.NewBufferString(invalidURL))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_CreateShortLink_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := new(MockService)
	h := handler.NewHandler(mockService)

	// Service returns an error
	mockService.On("CreateShortLink", mock.Anything, "https://example.com").Return("", errors.New("service error"))

	router.POST("/create", h.CreateShortLink)

	validURL := `{"long_url":"https://example.com"}`
	req, _ := http.NewRequest(http.MethodPost, "/create", bytes.NewBufferString(validURL))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert that the response code is 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_RedirectToLongURL(t *testing.T) {
	mockService := new(MockService)
	h := handler.NewHandler(mockService)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/:shortLink", h.RedirectToLongURL)

	shortLink := "testShortLink"
	longURL := "http://example.com"
	mockService.On("GetLongURL", mock.Anything, shortLink).Return(longURL, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/"+shortLink, nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, longURL, w.Header().Get("Location"))
	mockService.AssertExpectations(t)
}

func TestHandler_GetStats(t *testing.T) {
	mockService := new(MockService)
	h := handler.NewHandler(mockService)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/stats/:shortLink", h.GetStats)

	shortLink := "abc123"
	stats := model.URL{
		ID:          100,
		LongURL:     "http://example.com",
		AccessCount: 42,
	}
	mockService.On("GetLinkStats", mock.Anything, shortLink).Return(&stats, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/stats/"+shortLink, nil)
	r.ServeHTTP(w, req)

	expectedBody := handler.GetStatsResponse{
		LongURL:     stats.LongURL,
		ShortLink:   shortLink,
		AccessCount: stats.AccessCount,
	}
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, expectedBody.LongURL, stats.LongURL)
	assert.Equal(t, expectedBody.ShortLink, shortLink)
	assert.Equal(t, expectedBody.AccessCount, stats.AccessCount)
	mockService.AssertExpectations(t)
}
