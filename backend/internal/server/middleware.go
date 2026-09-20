package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"github.com/o-ga09/go-backend-template/internal/database/mysql"
	"github.com/o-ga09/go-backend-template/pkg/constant"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	"github.com/o-ga09/go-backend-template/pkg/errors"
	"github.com/o-ga09/go-backend-template/pkg/session"
	"github.com/o-ga09/go-backend-template/pkg/uuid"
)

type RequestInfo struct {
	status                                            int
	contents_length                                   int64
	method, path, sourceIP, query, user_agent, errors string
	elapsed                                           time.Duration
}

func AddID(ctx context.Context) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx := Ctx.SetRequestID(ctx, uuid.GenerateID())
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func WithTimeout() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
			defer cancel()
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			req := c.Request()
			slog.Log(req.Context(), constant.SeverityInfo, "処理開始", "request Id", Ctx.GetRequestID(req.Context()))

			err := next(c)

			status := http.StatusOK
			if res, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil {
				status = res.Status
			}
			r := &RequestInfo{
				status:          status,
				contents_length: req.ContentLength,
				method:          req.Method,
				path:            req.URL.Path,
				sourceIP:        c.RealIP(),
				query:           req.URL.RawQuery,
				user_agent:      req.UserAgent(),
				errors:          "",
				elapsed:         time.Since(start),
			}
			if err != nil {
				r.errors = err.Error()
			}
			slog.Log(req.Context(), constant.SeverityInfo, "処理終了", "Request", r.LogValue(), "requestId", Ctx.GetRequestID(req.Context()))
			return err
		}
	}
}

func (r *RequestInfo) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("status", r.status),
		slog.Int64("Content-length", r.contents_length),
		slog.String("method", r.method),
		slog.String("path", r.path),
		slog.String("sourceIP", r.sourceIP),
		slog.String("query", r.query),
		slog.String("user_agent", r.user_agent),
		slog.String("errors", r.errors),
		slog.String("elapsed", r.elapsed.String()),
	)
}

// CORS はcrossOriginでのCookie送信(credentials: 'include')を許可するCORS設定。
// フロントエンド(Next.js)はセッションCookieを使うため、ワイルドカードオリジン
// (AllowOrigins: []string{"*"})とAllowCredentials: trueの組み合わせは
// Fetch仕様上ブラウザに拒否される。特定オリジン(pkg/config.Config.FrontendOrigin)
// のみを許可する。
func CORS(ctx context.Context) echo.MiddlewareFunc {
	cfg := Ctx.GetCfgFromCtx(ctx)
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{cfg.FrontendOrigin},
		AllowMethods: []string{
			http.MethodPost,
			http.MethodGet,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours in seconds
	})
}

// Authenticate はセッションCookie(pkg/session.CookieName)を検証し、有効であれば
// ユーザーIDのみをcontextに格納する(pkg/context.SetUserID)。
// Cookieが無い/不正な場合でもリクエスト自体は拒否せずそのまま次へ進める。
// ログインを必須とするかどうかはハンドラ側がpkg/context.GetUserIDの結果を見て
// 判断する(context-propagation.md「認可ミドルウェア／ドメイン層で明示的に検証する」)。
func Authenticate(mgr *session.Manager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx := c.Request().Context()

			cookie, err := c.Cookie(session.CookieName)
			if err == nil && cookie.Value != "" {
				if userID, verifyErr := mgr.Verify(cookie.Value); verifyErr == nil {
					ctx = Ctx.SetUserID(ctx, userID)
					c.SetRequest(c.Request().WithContext(ctx))
				}
			}
			return next(c)
		}
	}
}

func SetDB() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx := c.Request().Context()
			db, err := mysql.Connect(ctx)
			if err != nil {
				slog.Log(ctx, constant.SeverityError, "DB接続に失敗しました", "error", err.Error())
				return err
			}
			ctx = Ctx.SetDB(ctx, db)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func AddTime() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx := Ctx.SetRequestTime(c.Request().Context(), time.Now())
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func ErrorHandler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			err := next(c)
			if err != nil {
				status, message := ErrCodeToStatusAndMessage(err)
				return c.JSON(status, map[string]interface{}{
					"error": message,
				})
			}
			return nil
		}
	}
}

func ErrCodeToStatusAndMessage(err error) (int, string) {
	code := errors.GetCode(err)
	switch code {
	case errors.ErrCodeUnAuthorized:
		return http.StatusUnauthorized, code.Message()
	case errors.ErrCodeUnAuthorization:
		return http.StatusForbidden, code.Message()
	case errors.ErrCodeInvalidArgument:
		// error-handling.mdの表: バリデーション違反(MakeBusinessError)は422。
		return http.StatusUnprocessableEntity, code.Message()
	case errors.ErrCodeConflict:
		return http.StatusConflict, code.Message()
	case errors.ErrCodeNotFound:
		return http.StatusNotFound, code.Message()
	case errors.ErrCodeSystem:
		return http.StatusInternalServerError, code.Message()
	default:
		return http.StatusInternalServerError, "Internal Server Error"
	}
}
