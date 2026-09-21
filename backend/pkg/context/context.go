package context

import (
	"context"
	"time"

	"github.com/o-ga09/go-backend-template/pkg/config"
	"gorm.io/gorm"
)

type RequestId string
type RequestTime string
type DB string
type UserID string

const RequestIDKey RequestId = "requestId"
const RequestTimeKey RequestTime = "requestTime"
const DBKey DB = "db"

// UserIDKey はAuthenticateミドルウェアが検証したセッションの持ち主のユーザーIDを
// 格納するcontextキー。認証情報(トークン・パスワード等)や個人情報(メールアドレス等)
// は格納しない。値が入っていること自体を「認証済みの証明」として扱わず、
// 権限判定が必要な箇所ではドメイン層の判定関数(例: user.IsOwnedBy)に明示的に渡す
// (context-propagation.md参照)。
const UserIDKey UserID = "userId"

func GetRequestID(ctx context.Context) string {
	return ctx.Value(RequestIDKey).(string)
}

func SetRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, RequestIDKey, id)
}

func GetCfgFromCtx(ctx context.Context) *config.Config {
	return ctx.Value(config.ConfigKey).(*config.Config)
}

func SetRequestTime(ctx context.Context, reqTime time.Time) context.Context {
	return context.WithValue(ctx, RequestTimeKey, reqTime)
}

func GetRequestTime(ctx context.Context) time.Time {
	t, ok := ctx.Value(RequestTimeKey).(time.Time)
	if !ok {
		return time.Time{}
	}
	return t
}

func SetDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, DBKey, db)
}

func GetDBFromCtx(ctx context.Context) *gorm.DB {
	db, ok := ctx.Value(DBKey).(*gorm.DB)
	if !ok {
		return nil
	}
	return db.WithContext(ctx)
}

// SetUserID はセッション検証済みのユーザーIDをcontextに格納する。
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID はcontextからユーザーIDを取得する。未ログイン(値未設定)の場合は
// 空文字列を返す(RequestIDと異なりpanicしない。未認証リクエストが通常経路として
// 存在するため)。
func GetUserID(ctx context.Context) string {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}
