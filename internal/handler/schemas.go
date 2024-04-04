package handler

type CreateLinkRequest struct {
	LongURL string `json:"long_url" binding:"required,url"`
}

type GetStatsResponse struct {
	LongURL     string `json:"long_url"`
	ShortLink   string `json:"short_link"`
	AccessCount int64  `json:"access_count"`
}
