package service_test

import (
	"context"
	"shortlink-go/internal/model"
	"shortlink-go/internal/service"
	"shortlink-go/pkg/base62"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockURLRepository struct {
	mock.Mock
}

func (m *MockURLRepository) CreateShortLink(ctx context.Context, longURL string) (int64, error) {
	args := m.Called(ctx, longURL)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockURLRepository) GetLongURL(ctx context.Context, id int64) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

func (m *MockURLRepository) GetURLStats(ctx context.Context, id int64) (*model.URL, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*model.URL), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockURLRepository) IncrementAccessCount(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	if cmd := args.Get(0); cmd != nil {
		return cmd.(*redis.StatusCmd)
	}
	return nil
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func TestService_CreateShortLink(t *testing.T) {
	mockURLRepo := new(MockURLRepository)
	mockRedisClient := new(MockRedisClient)
	svc := service.NewService(mockURLRepo, mockRedisClient)

	ctx := context.Background()
	longURL := "http://example.com"
	mockID := int64(1)
	expectedShortLink := base62.Encode(mockID)

	mockURLRepo.On("CreateShortLink", ctx, longURL).Return(mockID, nil)
	mockRedisClient.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&redis.StatusCmd{})

	shortLink, err := svc.CreateShortLink(ctx, longURL)

	assert.NoError(t, err)
	assert.Equal(t, expectedShortLink, shortLink)
	mockURLRepo.AssertExpectations(t)
	mockRedisClient.AssertExpectations(t)
}

func TestService_GetLongURL_RedisHit(t *testing.T) {
	mockURLRepo := new(MockURLRepository)
	mockRedisClient := new(MockRedisClient)
	svc := service.NewService(mockURLRepo, mockRedisClient)

	ctx := context.Background()
	shortLink := "abc123"
	expectedLongURL := "http://example.com"
	mockRedisClient.On("Get", ctx, service.REDIS_KEY_PREFIX+shortLink).Return(redis.NewStringResult(expectedLongURL, nil))
	mockURLRepo.On("IncrementAccessCount", mock.Anything, mock.Anything).Return(nil)

	longURL, err := svc.GetLongURL(ctx, shortLink)

	assert.NoError(t, err)
	assert.Equal(t, expectedLongURL, longURL)
	mockRedisClient.AssertExpectations(t)
	mockURLRepo.AssertNotCalled(t, "GetLongURL", mock.Anything, mock.Anything)
}

func TestService_GetLongURL_RedisMiss_DBHit(t *testing.T) {
	mockURLRepo := new(MockURLRepository)
	mockRedisClient := new(MockRedisClient)
	svc := service.NewService(mockURLRepo, mockRedisClient)

	ctx := context.Background()
	shortLink := "abc123"
	expectedID := base62.Decode(shortLink)
	expectedLongURL := "http://example.com"
	mockRedisClient.On("Get", ctx, service.REDIS_KEY_PREFIX+shortLink).Return(redis.NewStringResult("", redis.Nil))
	mockURLRepo.On("GetLongURL", ctx, expectedID).Return(expectedLongURL, nil)
	mockRedisClient.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&redis.StatusCmd{})
	mockURLRepo.On("IncrementAccessCount", mock.Anything, mock.Anything).Return(nil)

	longURL, err := svc.GetLongURL(ctx, shortLink)

	assert.NoError(t, err)
	assert.Equal(t, expectedLongURL, longURL)
	mockRedisClient.AssertExpectations(t)
	mockURLRepo.AssertCalled(t, "GetLongURL", mock.Anything, expectedID)
}

func TestService_GetLinkStats_Success(t *testing.T) {
	mockURLRepo := new(MockURLRepository)
	svc := service.NewService(mockURLRepo, nil) // For this test, Redis interaction is not involved

	ctx := context.Background()
	shortLink := "abc123"
	decodedID := base62.Decode(shortLink)
	expectedStats := &model.URL{
		ID:          decodedID,
		LongURL:     "http://example.com",
		AccessCount: 42,
	}

	mockURLRepo.On("GetURLStats", ctx, decodedID).Return(expectedStats, nil)

	stats, err := svc.GetLinkStats(ctx, shortLink)

	assert.NoError(t, err)
	assert.Equal(t, expectedStats, stats)
	mockURLRepo.AssertExpectations(t)
}
