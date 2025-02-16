package server

import (
	"github.com/gin-gonic/gin"
	"github.com/o-ga09/go-backend-template/internal/config"
	"github.com/o-ga09/go-backend-template/internal/middleware"
	"github.com/o-ga09/go-backend-template/internal/service"
)

func NewServer() (*gin.Engine, error) {
	r := gin.New()
	cfg, _ := config.New()
	if cfg.Env == "PROD" {
		gin.SetMode(gin.ReleaseMode)
	}

	// ロガー設定
	logger := middleware.Logger()
	httpLogger := middleware.RequestLogger(logger)

	// CORS設定
	cors := middleware.CORS()

	// リクエストタイムアウト設定
	withCtx := middleware.WithTimeout()

	// リクエストID付与
	withReqId := middleware.AddID()

	// ミドルウェア設定
	r.Use(withReqId)
	r.Use(withCtx)
	r.Use(cors)
	r.Use(httpLogger)

	// ヘルスチェック
	v1 := r.Group("/v1")
	{
		systemHandler := service.SystemHandler{}
		v1.GET("/health", systemHandler.Health)

		// ユーザー一覧取得
		userHandler := service.UserHandler{}
		v1.GET("/users", userHandler.Find)
	}

	return r, nil
}
