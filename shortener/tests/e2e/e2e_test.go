package e2e

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"shortener/internal/service"
	"shortener/internal/storage"
	"shortener/internal/storage/postgres"
	"shortener/internal/storage/redis"
	"shortener/internal/web/handlers"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	DBHost     = "localhost"
	DBUser     = "test"
	DBPassword = "test"
	DBName     = "testdb"

	DBMapped    = "5432"
	RedisMapped = "6379"
)

func SetupTestDB(t *testing.T) testcontainers.Container {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:13-alpine",
		ExposedPorts: []string{fmt.Sprintf("%s/tcp", DBMapped)},
		Env: map[string]string{
			"POSTGRES_USER":     DBUser,
			"POSTGRES_PASSWORD": DBPassword,
			"POSTGRES_DB":       DBName,
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, _ := pgContainer.Host(ctx)
	port, _ := pgContainer.MappedPort(ctx, DBMapped)
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port.Port(), DBUser, DBPassword, DBName,
	)

	time.Sleep(time.Second * 3)

	oldDB, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	defer func() {
		_ = oldDB.Close()
	}()

	err = oldDB.Ping()
	require.NoError(t, err)

	applyMigrations(t, oldDB)

	return pgContainer
}

func applyMigrations(t *testing.T, db *sql.DB) {
	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "../../migrations")

	goose.SetBaseFS(nil)

	if err := goose.Up(db, migrationsDir); err != nil {
		t.Fatalf("Failed to apply migrations: %v", err)
	}

}

func SetupTestRedis(t *testing.T) testcontainers.Container {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "redis:7.2-alpine",
		ExposedPorts: []string{fmt.Sprintf("%s/tcp", RedisMapped)},
		Env: map[string]string{
			"MAXMEMORY":        "100MB",
			"MAXMEMORY_POLICY": "volatile-ttl",
		},
		WaitingFor: wait.ForLog("Ready to accept connections"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	time.Sleep(time.Second * 3)

	return redisContainer
}

func TestMain(t *testing.T) {

	gin.SetMode(gin.TestMode)
	g := gin.Default()

	dbCont := SetupTestDB(t)
	defer func() { _ = dbCont.Terminate(context.Background()) }()

	dbHost, err := dbCont.Host(context.Background())
	require.NoError(t, err)
	dbPort, err := dbCont.MappedPort(context.Background(), DBMapped)
	require.NoError(t, err)

	db := postgres.New(dbHost, dbPort.Port(), DBUser, DBPassword, DBName, "disable")

	rdCont := SetupTestRedis(t)
	defer func() { _ = rdCont.Terminate(context.Background()) }()

	rdHost, err := rdCont.Host(context.Background())
	require.NoError(t, err)
	rdPort, err := rdCont.MappedPort(context.Background(), RedisMapped)
	require.NoError(t, err)

	rd := redis.New(fmt.Sprintf("%s:%s", rdHost, rdPort.Port()), "", 0)

	str := storage.New(db, rd)
	srv := service.New(str, str)

	// ---------------- CHECK CREATING -------------------
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost, "/endpoint",
		strings.NewReader(`{"original": "https://vk.com"}`),
	)

	g.POST("/endpoint", handlers.NewShort(srv))

	g.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Result().StatusCode)
	type res struct {
		Result string `json:"result"`
	}
	var alias res
	err = json.Unmarshal(rr.Body.Bytes(), &alias)
	require.NoError(t, err)

	short, err := db.URL(alias.Result)
	require.NoError(t, err)
	require.NotEqual(t, "", short)
	// ---------------------------------------------------

	// --------------------- CHECK GETTING ---------------
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(
		http.MethodGet, fmt.Sprintf("/getting/%s", alias.Result), nil,
	)
	g.GET("/getting/:short_url", handlers.Redirect(srv))
	g.ServeHTTP(rr, req)
	require.Equal(t, http.StatusTemporaryRedirect, rr.Result().StatusCode)
	cached, err := rd.URL(alias.Result)
	require.NoError(t, err)
	require.NotEqual(t, "", cached)
	// ---------------------------------------------------

	srv.Shutdown()
	str.Shutdown()
}
