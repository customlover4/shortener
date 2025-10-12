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

type AgrigateOpts struct {
	Alias          string
	StartDate      string `form:"start_date"`
	EndDate        string `form:"end_date"`
	FilterColumn   string `form:"filter"`
	ValueForFilter string `form:"value"`
	Page           int    `form:"page"`
}
