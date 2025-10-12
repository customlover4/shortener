package web

import (
	"shortener/internal/service"
	"shortener/internal/web/handlers"

	f "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/wb-go/wbf/ginext"
)

func SetRoutes(r *ginext.Engine, s *service.Service) {
	r.GET("/s/:short_url", handlers.Redirect(s))
	r.POST("/shorten", handlers.NewShort(s))
	r.GET("/analytics/:short_url", handlers.Analytics(s))

	r.Static("/static", "./templates/static")

	r.GET(
		"/swagger/*any", ginSwagger.WrapHandler(f.Handler),
	)

	r.GET("/", handlers.MainHandler())

}
