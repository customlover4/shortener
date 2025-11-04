package handlers

import (
	"errors"
	"net/http"
	urlParser "net/url"
	"shortener/internal/entities/redirect"
	"shortener/internal/entities/request"
	"shortener/internal/entities/response"
	"shortener/internal/entities/url"
	"shortener/internal/service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

type servicer interface {
	// for urls
	CreateURL(u url.URL) (string, error)
	URL(alias string) (string, error)

	// for redirects
	CreateRedirect(redirects redirect.Redirect)
	Redirects(alias string) ([]redirect.Redirect, error)
	AgrigatedRedirects(opts redirect.AgrigateOpts) (redirect.Agrigated, error)
}

func MainHandler() gin.HandlerFunc {
	return func(ctx *ginext.Context) {
		ctx.HTML(200, "main.html", nil)
	}
}

// NewShort creates a new URL alias.
// @Summary Create a new short URL alias
// @Description Creates a short alias for a given original URL.
// @Tags URLs
// @Accept json
// @Produce json
// @Param request body request.NewShort true "Shortening request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure 503 {object} response.Response
// @Router /shorten [post]
func NewShort(s servicer) gin.HandlerFunc {
	return func(ctx *ginext.Context) {
		const op = "internal.handlers.newShort"

		var ns request.NewShort
		if err := ctx.BindJSON(&ns); err != nil {
			ctx.JSONP(http.StatusBadRequest, response.Error(
				"wrong json values (type)",
			))
			return
		}
		short, msg := ns.Validate()
		if msg != "" {
			ctx.JSONP(http.StatusBadRequest, response.Error(
				msg,
			))
			return
		}

		alias, err := s.CreateURL(short)
		if errors.Is(err, service.ErrNotValidData) {
			ctx.JSONP(http.StatusServiceUnavailable, response.Error(
				"url format: http://example.com",
			))
			return
		} else if errors.Is(err, service.ErrNotUnique) {
			ctx.JSONP(http.StatusServiceUnavailable, response.Error(
				"not unique alias",
			))
			return
		} else if err != nil {
			zlog.Logger.Error().Err(err).Msg("op: " + op)
			ctx.JSONP(http.StatusInternalServerError, response.Error(
				"internal server error on our service",
			))
			return
		}

		ctx.JSONP(http.StatusOK, response.OK(alias))
	}
}

// Redirect to original URL by alias.
// @Summary Redirect by alias
// @Description Redirects user to the original URL. On error, returns JSON.
// @Tags URLs
// @Param alias path string true "Short URL alias"
// @Success 307 {string} string "Temporary redirect to original URL"
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure 503 {object} response.Response
// @Router /s/{alias} [get]
func Redirect(s servicer) gin.HandlerFunc {
	return func(ctx *ginext.Context) {
		const op = "internal.handlers.Redirect"

		alias := ctx.Param("short_url")
		original, err := s.URL(alias)
		if errors.Is(err, service.ErrNotValidData) {
			ctx.JSONP(http.StatusServiceUnavailable, response.Error(
				"not valid data for redirecting",
			))
			return
		} else if errors.Is(err, service.ErrNotFound) {
			ctx.JSONP(http.StatusNotFound, response.Error(
				"not found link",
			))
			return
		} else if err != nil {
			zlog.Logger.Error().Err(err).Msg("op: " + op)
			ctx.JSONP(http.StatusInternalServerError, response.Error(
				"internal server error on our service",
			))
			return
		}

		go s.CreateRedirect(redirect.Redirect{
			Alias:     alias,
			Date:      time.Now().UTC(),
			UserAgent: ctx.Request.UserAgent(),
		})

		url, err := urlParser.Parse(original)
		if err != nil {
			zlog.Logger.Error().Err(err).Msg("op: " + op)
			ctx.JSONP(http.StatusInternalServerError, response.Error(
				"internal server error on our service",
			))
			return
		}
		if url.Scheme == "" {
			url.Scheme = "https"
		}

		ctx.Redirect(http.StatusTemporaryRedirect, url.String())
	}
}

// GetAnalytics returns redirect analytics in JSON format.
// @Summary Get redirect analytics (JSON API)
// @Description Fetches aggregated redirect data for a short URL alias with filters and pagination.
// @Tags Analytics
// @Accept json
// @Produce json
// @Param alias path string true "Short URL alias"
// @Param start_date query string true "Start date in ISO8601 format" default(2025-01-01T00:00:00Z)
// @Param end_date query string true "End date in ISO8601 format" default(2025-12-31T23:59:59Z)
// @Param filter query string false "Column to filter by (e.g. user_agent)" default(user_agent)
// @Param value query string false "Value to filter" default(Mozilla/5.0...)
// @Param page query integer false "Page number" default(1)
// @Success 200 {object} redirect.Agrigated
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /analytics/{alias} [get]
func Analytics(s servicer) gin.HandlerFunc {
	return func(ctx *ginext.Context) {
		const op = "internal.handlers.Redirects"
		var opts redirect.AgrigateOpts
		if err := ctx.ShouldBindQuery(&opts); err != nil {
			ctx.HTML(http.StatusBadRequest, "400.html", struct{ Msg string }{err.Error()})
			return
		}

		alias := ctx.Param("short_url")
		opts.Alias = alias

		redirects, err := s.AgrigatedRedirects(opts)
		if errors.Is(err, service.ErrNotFound) {
			ctx.HTML(http.StatusNotFound, "404.html", nil)
			return
		} else if errors.Is(err, service.ErrNotValidData) {
			ctx.HTML(http.StatusServiceUnavailable, "400.html", struct{ Msg string }{err.Error()})
			return
		} else if err != nil {
			zlog.Logger.Error().Err(err).Msg("op: " + op)
			ctx.HTML(http.StatusInternalServerError, "500.html", nil)
			return
		}

		if opts.Page == 0 {
			opts.Page++
		}
		ctx.HTML(http.StatusOK, "redirects.html", gin.H{
			"Aggregated": redirects,
			"Opts":       opts,
		})
	}
}
