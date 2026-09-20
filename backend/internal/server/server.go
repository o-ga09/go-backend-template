package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/o-ga09/go-backend-template/internal/database/mysql"
	"github.com/o-ga09/go-backend-template/internal/handler"
	"github.com/o-ga09/go-backend-template/internal/router"
	"github.com/o-ga09/go-backend-template/pkg/binder"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	"github.com/o-ga09/go-backend-template/pkg/logger"
	"github.com/o-ga09/go-backend-template/pkg/session"
	"github.com/o-ga09/go-backend-template/pkg/validator"
)

// sessionTTL はバックエンド発行セッションCookieの有効期間(backend/docs/auth.md参照)。
const sessionTTL = 7 * 24 * time.Hour

type Server struct {
	Port   string
	engine *echo.Echo
}

func New(ctx context.Context) *Server {
	cfg := Ctx.GetCfgFromCtx(ctx)
	engine := echo.New()
	engine.Validator = validator.New()
	engine.Binder = binder.New()
	return &Server{
		Port:   cfg.Port,
		engine: engine,
	}
}

func (s *Server) Run(ctx context.Context) error {
	cfg := Ctx.GetCfgFromCtx(ctx)
	sessionMgr := session.NewManager(cfg.SessionSecret, sessionTTL)

	// 依存の組み立て(コンストラクタで注入。architecture.md「依存注入のルール」)
	userRepo := mysql.NewUserRepository()
	userHandler := handler.NewUserHandler(userRepo, sessionMgr)
	authHandler := handler.NewAuthHandler(userRepo)

	// ミドルウェアの設定
	s.engine.Use(middleware.Recover())
	s.engine.Use(AddID(ctx))
	s.engine.Use(AddTime())
	s.engine.Use(RequestLogger())
	s.engine.Use(SetDB())
	s.engine.Use(WithTimeout())
	s.engine.Use(CORS(ctx))
	s.engine.Use(Authenticate(sessionMgr))
	s.engine.Use(middleware.BodyLimit("10M"))
	s.engine.Use(middleware.Gzip())
	s.engine.Use(ErrorHandler())

	// ルーティングの設定
	apiRoot := s.engine.Group("/api")
	router.SetupApplicationRoute(apiRoot, userHandler, authHandler)
	router.SetupSystemRoute(apiRoot)
	// サーバーの起動
	port := fmt.Sprintf(":%s", s.Port)
	srv := &http.Server{
		Addr:    port,
		Handler: s.engine,
	}

	// サーバーの起動
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, fmt.Sprintf("Failed to listen and serve: %v", err))
		}
	}()

	logger.Info(ctx, fmt.Sprintf("Server is running on %s", s.Port))
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info(ctx, "graceful shutdown")

	// サーバーのタイムアウト設定
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// サーバーのシャットダウン
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to shutdown server: %v", err))
		return err
	}
	return nil
}
