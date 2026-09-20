package config

import (
	"context"

	"github.com/caarlos0/env"
)

type CfgKey string

const ConfigKey CfgKey = "config"

type Config struct {
	Env          string `env:"ENV" envDefault:"dev"`
	Port         string `env:"PORT" envDefault:"80"`
	Database_url string `env:"DATABASE_URL" envDefult:""`
	ProjectID    string `env:"PROJECTID" envDefault:""`

	// FrontendOrigin はCORSで許可するフロントエンド(Next.js)のオリジン。
	// フロントエンドはCookie(credentials: 'include')でセッションを送るため、
	// AllowOrigins にはワイルドカードではなく単一の具体的なオリジンを設定する
	// 必要がある(internal/server/middleware.go の CORS() 参照)。
	FrontendOrigin string `env:"FRONTEND_ORIGIN" envDefault:"http://localhost:3000"`

	// SessionSecret はバックエンド発行セッションCookieの署名鍵(HMAC-SHA256)。
	// 本番環境では必ず十分なエントロピーを持つ値を設定すること
	// (backend/docs/auth.md参照)。
	SessionSecret string `env:"SESSION_SECRET" envDefault:"dev-insecure-session-secret-change-me"`
}

func New(ctx context.Context) (context.Context, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return context.WithValue(ctx, ConfigKey, cfg), nil
}
