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

const RequestIDKey RequestId = "requestId"
const RequestTimeKey RequestTime = "requestTime"
const DBKey DB = "db"

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
	return db
}
