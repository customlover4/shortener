package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"shortener/internal/entities/redirect"
	"shortener/internal/entities/url"
	"shortener/internal/service"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type serviceMock struct {
	createURLF func(u url.URL) (string, error)
	getURLF    func(alias string) (string, error)

	createRedirectF func(r redirect.Redirect)
	getRedirectsF   func(alias string) ([]redirect.Redirect, error)
	agrigatedF      func(opts redirect.AgrigateOpts) (redirect.Agrigated, error)
}

func (sm *serviceMock) CreateURL(u url.URL) (string, error) {
	return sm.createURLF(u)
}
func (sm *serviceMock) URL(alias string) (string, error) {
	return sm.getURLF(alias)
}

func (sm *serviceMock) CreateRedirect(r redirect.Redirect) {
	sm.createRedirectF(r)
}
func (sm *serviceMock) Redirects(alias string) ([]redirect.Redirect, error) {
	return sm.getRedirectsF(alias)
}

func (sm *serviceMock) AgrigatedRedirects(opts redirect.AgrigateOpts) (redirect.Agrigated, error) {
	return sm.agrigatedF(opts)
}

func TestNewShort(t *testing.T) {
	type args struct {
		servicer servicer
	}
	tests := []struct {
		name string
		body string
		args args
		want int
	}{
		{
			name: "good",
			body: `{"original": "https://test.com"}`,
			args: args{
				servicer: &serviceMock{
					createURLF: func(u url.URL) (string, error) {
						return "ok", nil
					},
				},
			},
			want: http.StatusOK,
		},
		{
			name: "bad json",
			body: `{"original": 123}`,
			args: args{
				servicer: &serviceMock{
					createURLF: func(u url.URL) (string, error) {
						return "ok", nil
					},
				},
			},
			want: http.StatusBadRequest,
		},
		{
			name: "not valid json",
			body: `{"alias": "hihi"}`,
			args: args{
				servicer: &serviceMock{
					createURLF: func(u url.URL) (string, error) {
						return "ok", nil
					},
				},
			},
			want: http.StatusBadRequest,
		},
		{
			name: "not unique alias",
			body: `{"alias": "test", "original": "https://test.com"}`,
			args: args{
				servicer: &serviceMock{
					createURLF: func(u url.URL) (string, error) {
						return "ok", service.ErrNotUnique
					},
				},
			},
			want: http.StatusServiceUnavailable,
		},
		{
			name: "not valid data",
			body: `{"original": "test.com"}`,
			args: args{
				servicer: &serviceMock{
					createURLF: func(u url.URL) (string, error) {
						return "ok", service.ErrNotValidData
					},
				},
			},
			want: http.StatusServiceUnavailable,
		},
		{
			name: "storage internal",
			body: `{"original": "https://test.com"}`,
			args: args{
				servicer: &serviceMock{
					createURLF: func(u url.URL) (string, error) {
						return "ok", errors.New("unknown")
					},
				},
			},
			want: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodPost, "/endpoint", strings.NewReader(tt.body),
			)
			router := gin.Default()
			router.POST("/endpoint", NewShort(tt.args.servicer))
			router.ServeHTTP(rr, req)
			if rr.Result().StatusCode != tt.want {
				t.Errorf(
					"NewShort() status code get=%d, want %d",
					rr.Result().StatusCode, tt.want,
				)
			}
		})
	}
}

func TestRedirect(t *testing.T) {
	type args struct {
		servicer servicer
	}
	tests := []struct {
		name  string
		alias string
		args  args
		want  int
	}{
		{
			name:  "good",
			alias: "alias",
			args: args{
				servicer: &serviceMock{
					getURLF: func(alias string) (string, error) {
						return "original", nil
					},
					createRedirectF: func(r redirect.Redirect) {
					},
				},
			},
			want: http.StatusTemporaryRedirect,
		},
		{
			name:  "not valid data",
			alias: "jhjkhjjkhjkhkj",
			args: args{
				servicer: &serviceMock{
					getURLF: func(alias string) (string, error) {
						return "original", service.ErrNotValidData
					},
					createRedirectF: func(r redirect.Redirect) {
					},
				},
			},
			want: http.StatusServiceUnavailable,
		},
		{
			name:  "not found alias",
			alias: "jhjkhjjkhjkhkj",
			args: args{
				servicer: &serviceMock{
					getURLF: func(alias string) (string, error) {
						return "original", service.ErrNotFound
					},
					createRedirectF: func(r redirect.Redirect) {
					},
				},
			},
			want: http.StatusNotFound,
		},
		{
			name:  "internal errpr",
			alias: "jhjkhjjkhjkhkj",
			args: args{
				servicer: &serviceMock{
					getURLF: func(alias string) (string, error) {
						return "original", errors.New("unknown")
					},
					createRedirectF: func(r redirect.Redirect) {
					},
				},
			},
			want: http.StatusInternalServerError,
		},
		{
			name:  "cant parse original",
			alias: "jhjkhjjkhjkhkj",
			args: args{
				servicer: &serviceMock{
					getURLF: func(alias string) (string, error) {
						return "9*@&(&$%())", nil
					},
					createRedirectF: func(r redirect.Redirect) {
					},
				},
			},
			want: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodGet, "/endpoint/"+tt.alias, nil,
			)
			router := gin.Default()
			router.GET("/endpoint/:short_url", Redirect(tt.args.servicer))
			router.ServeHTTP(rr, req)
			if rr.Result().StatusCode != tt.want {
				t.Errorf(
					"Redirect() status code get=%d, want %d",
					rr.Result().StatusCode, tt.want,
				)
			}
		})
	}
}

func TestAnalytics(t *testing.T) {
	type args struct {
		servicer servicer
	}
	tests := []struct {
		name  string
		alias string
		args  args
		query string
		want  int
	}{
		{
			name:  "good",
			alias: "Test",
			args: args{
				servicer: &serviceMock{
					agrigatedF: func(opts redirect.AgrigateOpts) (redirect.Agrigated, error) {
						return redirect.Agrigated{}, nil
					},
				},
			},
			want: http.StatusOK,
		},
		{
			name:  "bad query",
			alias: "Test",
			args: args{
				servicer: &serviceMock{
					agrigatedF: func(opts redirect.AgrigateOpts) (redirect.Agrigated, error) {
						return redirect.Agrigated{}, nil
					},
				},
			},
			query: "page=how",
			want:  http.StatusBadRequest,
		},
		{
			name:  "not found alias",
			alias: "notfound",
			args: args{
				servicer: &serviceMock{
					agrigatedF: func(opts redirect.AgrigateOpts) (redirect.Agrigated, error) {
						return redirect.Agrigated{}, service.ErrNotFound
					},
				},
			},
			query: "page=1",
			want:  http.StatusNotFound,
		},
		{
			name:  "not valid data",
			alias: "notfound",
			args: args{
				servicer: &serviceMock{
					agrigatedF: func(opts redirect.AgrigateOpts) (redirect.Agrigated, error) {
						return redirect.Agrigated{}, service.ErrNotValidData
					},
				},
			},
			query: "page=1",
			want:  http.StatusServiceUnavailable,
		},
		{
			name:  "unknown error",
			alias: "notfound",
			args: args{
				servicer: &serviceMock{
					agrigatedF: func(opts redirect.AgrigateOpts) (redirect.Agrigated, error) {
						return redirect.Agrigated{}, errors.New("unknown")
					},
				},
			},
			query: "page=1",
			want:  http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			tmpDir := t.TempDir()

			templateContent := `<html><head></head></html>`

			templatePath := filepath.Join(tmpDir, "test.html")
			err := os.WriteFile(templatePath, []byte(templateContent), 0644)
			if err != nil {
				t.Errorf("can't create tmp dir with templates")
				return
			}

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodGet, "/endpoint/"+tt.alias+"?"+tt.query, nil,
			)
			router := gin.Default()
			router.GET("/endpoint/:short_url", Analytics(tt.args.servicer))
			router.LoadHTMLFiles(templatePath)
			router.ServeHTTP(rr, req)
			if rr.Result().StatusCode != tt.want {
				t.Errorf(
					"Analytics() status code get=%d, want %d",
					rr.Result().StatusCode, tt.want,
				)
			}
		})
	}
}
