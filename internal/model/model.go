package model

// URL struct represents the URL table structure from your database in Go.
type URL struct {
	ID          int64
	LongURL     string
	AccessCount int64
}
