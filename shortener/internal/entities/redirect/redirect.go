package redirect

import "time"

type Redirect struct {
	ID        int64
	Alias     string
	Date      time.Time
	UserAgent string
}

const (
	PageSize = 20
)

var (
	FilterUserAgent = "user_agent"
)

type Agrigated struct {
	Alias     string
	Total     int64
	Redirects []Redirect
}

// AgrigateOpts options for filtering and paginating redirect analytics.
type AgrigateOpts struct {
	// Alias of the short URL (required, from URL path)
	Alias string `json:"alias"`

	// StartDate in ISO8601 format, e.g. "2025-01-01T00:00:00Z"
	// Required.
	StartDate string `json:"start_date" form:"start_date" example:"2025-01-01T00:00:00Z"`

	// EndDate in ISO8601 format, e.g. "2025-12-31T23:59:59Z"
	// Required.
	EndDate string `json:"end_date" form:"end_date" example:"2025-12-31T23:59:59Z"`

	// FilterColumn name to filter by (e.g. "user_agent", "country")
	FilterColumn string `json:"filter" form:"filter" example:"user_agent"`

	// ValueForFilter value to match in the specified column
	ValueForFilter string `json:"value" form:"value" example:"Mozilla/5.0..."`

	// Page number for pagination (starts from 1)
	Page int `json:"page" form:"page" example:"1" default:"1"`
}
