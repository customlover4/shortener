package request

import "shortener/internal/entities/url"

type NewShort struct {
	Alias    string `json:"alias"`
	Original string `json:"original"`
}

func (ns NewShort) Validate() (url.URL, string) {
	var u url.URL
	if ns.Original == "" {
		return u, "empty original link"
	}
	u.Alias = ns.Alias
	u.Original = ns.Original
	return u, ""
}
