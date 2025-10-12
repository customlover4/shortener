package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"shortener/internal/service"
	"shortener/internal/storage"
	"shortener/internal/storage/postgres"
	"shortener/internal/storage/redis"
	"shortener/internal/web"
	"strconv"
	"syscall"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"

	"net/http"
)

// @title Shortener
// @version 0.0.1
// @description Создает алиасы для ссылок
// @host localhost
// @BasePath /

var (
	ConfigPath       = "../config/config.yml" // prod: os.Getenv("CONFIG_PATH")
	Port             = "80"                   // prod: os.Getenv("PORT")
	Templates        = "templates/*.html"     // prod: os.Getenv("templates")
	PostgresPassword = "qqq"                  // prod: os.Getenv("POSTGRES_PASSWORD")
	RedisPassword    = "qqq"                  // prod: os.Getenv("REDIS_PASSWORD")
)

func templates(router *ginext.Engine) {
	router.SetFuncMap(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
	})
	router.LoadHTMLGlob(Templates)
}

func init() {
	if os.Getenv("DEBUG") == "false" {
		gin.SetMode(gin.ReleaseMode)
	}

	ConfigPath = os.Getenv("CONFIG_PATH")
	Port = os.Getenv("PORT")
	Templates = os.Getenv("TEMPLATES")
	PostgresPassword = os.Getenv("POSTGRES_PASSWORD")
	RedisPassword = os.Getenv("REDIS_PASSWORD")
}

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	zlog.Init()

	cfg := config.New()
	err := cfg.Load(ConfigPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	rdI, err := strconv.Atoi(cfg.GetString("redis.db"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	rd := redis.New(
		cfg.GetString("redis.addr"),
		RedisPassword,
		rdI,
	)
	db := postgres.New(
		cfg.GetString("postgres.host"), cfg.GetString("postgres.port"),
		cfg.GetString("postgres.username"), PostgresPassword,
		cfg.GetString("postgres.dbname"), cfg.GetString("postgres.sslmode"),
	)
	str := storage.New(db, rd)
	srv := service.New(str, str)

	router := ginext.New()
	templates(router)
	web.SetRoutes(router, srv)
	server := &http.Server{
		Addr:           ":" + Port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	go func() {
		err := server.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}()

	zlog.Logger.Info().Msg("start listening port")
	<-sig
	zlog.Logger.Info().Msg("shutdown staring...")
	// start shutdown from higher layer to lower
	// http -> service -> storage
	server.Close()
	srv.Shutdown()
	str.Shutdown()
}
