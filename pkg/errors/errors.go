package errors

import (
	"context"
	"errors"

	"github.com/newmo-oss/ergo"
	"github.com/o-ga09/go-backend-template/pkg/logger"
)

type ErrType string

var (
	ErrTypeUnAuthorized    = ergo.NewSentinel("unauthorized")
	ErrTypeUnAuthorization = ergo.NewSentinel("unauthorization")
	ErrTypeBussiness       = ergo.NewSentinel("business error")
	ErrTypeConflict        = ergo.NewSentinel("conflict")
	ErrTypeNotFound        = ergo.NewSentinel("not found")
	ErrTypeCritical        = ergo.NewSentinel("critical error")
)

// TODO: 適宜プロジェクトごとに修正すること
var (
	// ドメインエラー
	ErrInvalidFirebaseID  = ergo.New("不正なFirebaseIDです。")
	ErrInvalidUserID      = ergo.New("不正なUserIDです。")
	ErrInvalidName        = ergo.New("不正なユーザー名です。")
	ErrInvalidDisplayName = ergo.New("不正な表示名です。")
	ErrInvalidGroupID     = ergo.New("不正なグループIDです。")
	ErrInvalidRelationID  = ergo.New("不正なリレーションIDです。")
	ErrInvalidTwitterID   = ergo.New("不正なTwitterIDです。")
	ErrInvalidGender      = ergo.New("性別の値の範囲が不正です。")
	ErrInvalidDateTime    = ergo.New("日付のフォーマットが不正です。")
	ErrInvalidProfileURL  = ergo.New("不正なプロフィールURLです。")
	ErrInvalidUserType    = ergo.New("無効なユーザータイプフォーマットです。")
	ErrFollowed           = ergo.New("すでにフォロー済みです。")
	ErrFollowSelf         = ergo.New("自分自身をフォローすることはできません。")
	ErrRequestNotNil      = ergo.New("リクエストが正しくありません。")

	// ulidエラー
	ErrEmptyULID   = ergo.New("empty ulid")
	ErrInvalidULID = ergo.New("invalid ulid")

	// データベースエラー
	ErrRecordNotFound         = ergo.New("record not found")
	ErrConflict               = ergo.New("conflict")
	ErrOptimisticLockConflict = ergo.New("optimistic lock conflict")
	ErrForeignKeyConstraint   = ergo.New("foreign key constraint error")
	ErrUniqueConstraint       = ergo.New("unique constraint error")

	// 画像エラー
	ErrInvalidImageType  = ergo.New("ファイルの種類が不正です。")
	ErrFailedImageName   = ergo.New("ファイル名の生成に失敗しました。")
	ErrFailedDecodeImage = ergo.New("画像のデコードに失敗しました。")
	ErrNotFoundImage     = ergo.New("画像が見つかりません。")

	// リクエストエラー
	ErrRequestBodyNil = ergo.New("リクエストボディが空です。")

	// その他エラー
	ErrSystem           = ergo.New("システムエラーが発生しました。")
	ErrAuthorized       = ergo.New("認証に失敗しました。")
	ErrUnauthorized     = ergo.New("認可に失敗しました。")
	ErrInvalidArgument  = ergo.New("バリデーションエラーが発生しました。")
	ErrInvalidOperation = ergo.New("無効な操作です。")
	ErrNotFound         = ergo.New("指定されたデータが見つかりません。")
)

func New(msg string) error {
	return ergo.New(msg)
}

func IsWrapped(err error) bool {
	return errors.Is(err, ErrTypeBussiness) ||
		errors.Is(err, ErrTypeUnAuthorized) ||
		errors.Is(err, ErrTypeUnAuthorization) ||
		errors.Is(err, ErrTypeConflict) ||
		errors.Is(err, ErrTypeNotFound) ||
		errors.Is(err, ErrTypeCritical)
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}

func GetMessage(err error) string {
	return err.Error()
}

func GetCode(err error) ergo.Code {
	return ergo.CodeOf(err)
}

func MakeAuthorizationError(ctx context.Context, msg string) error {
	err := ergo.Wrap(ErrUnauthorized, msg)
	st := ergo.StackTraceOf(err)
	logger.Warn(ctx, err.Error(), "callStack", st)
	return err
}

func MakeAuthorizedError(ctx context.Context, msg string) error {
	err := ergo.Wrap(ErrAuthorized, msg)
	st := ergo.StackTraceOf(err)
	logger.Warn(ctx, err.Error(), "callStack", st)
	return err
}

func MakeBusinessError(ctx context.Context, msg string) error {
	err := ergo.Wrap(ErrTypeBussiness, msg)
	st := ergo.StackTraceOf(err)
	logger.Info(ctx, err.Error(), "callStack", st)
	return err
}

func MakeConflictError(ctx context.Context, msg string) error {
	err := ergo.Wrap(ErrTypeConflict, msg)
	st := ergo.StackTraceOf(err)
	logger.Warn(ctx, err.Error(), "callStack", st)
	return err
}

func MakeNotFoundError(ctx context.Context, msg string) error {
	err := ergo.Wrap(ErrTypeNotFound, msg)
	st := ergo.StackTraceOf(err)
	logger.Warn(ctx, err.Error(), "callStack", st)
	return err
}

func MakeSystemError(ctx context.Context, err error) error {
	if !IsWrapped(err) {
		err = ergo.Wrap(ErrTypeCritical, err.Error())
	}
	st := ergo.StackTraceOf(err)
	logger.Error(ctx, err.Error(), "callStack", st)
	return err
}
