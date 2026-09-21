package errors

import (
	"context"
	"errors"

	"github.com/newmo-oss/ergo"
	"gorm.io/gorm"

	"github.com/o-ga09/go-backend-template/pkg/logger"
)

type ErrType string
type ErrCode ergo.Code

var (
	ErrTypeUnAuthorized    = ergo.NewSentinel("unauthorized")
	ErrTypeUnAuthorization = ergo.NewSentinel("unauthorization")
	ErrTypeBussiness       = ergo.NewSentinel("business error")
	ErrTypeConflict        = ergo.NewSentinel("conflict")
	ErrTypeNotFound        = ergo.NewSentinel("not found")
	ErrTypeCritical        = ergo.NewSentinel("critical error")
)

var (
	ErrCodeInvalidArgument = ergo.NewCode("422", "InvalidArgument")
	ErrCodeUnAuthorized    = ergo.NewCode("401", "UnAuthorized")
	ErrCodeUnAuthorization = ergo.NewCode("403", "UnAuthorization")
	ErrCodeNotFound        = ergo.NewCode("404", "NotFound")
	ErrCodeConflict        = ergo.NewCode("409", "Conflict")
	ErrCodeSystem          = ergo.NewCode("500", "SystemError")
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
	// ErrRecordNotFound/ErrForeignKeyConstraint/ErrUniqueConstraintはGORMの
	// センチネルそのもの。TranslateError:trueなGORM接続(mysql/connect.go)が
	// ドライバ固有のエラー番号(1062/1451/1452等)をこれらに変換するため、
	// mysqlパッケージのリポジトリは個別に変換せずそのまま返す(coding-style.md参照)。
	ErrRecordNotFound         = gorm.ErrRecordNotFound
	ErrConflict               = ergo.New("conflict")
	ErrOptimisticLockConflict = ergo.New("optimistic lock conflict")
	ErrForeignKeyConstraint   = gorm.ErrForeignKeyViolated
	ErrUniqueConstraint       = gorm.ErrDuplicatedKey

	// セッションエラー
	ErrInvalidSession = ergo.NewSentinel("invalid session")
	ErrSessionExpired = ergo.NewSentinel("session expired")

	// カートエラー
	// ErrCartEmpty はカートに商品が無い状態で注文確定しようとした場合のエラー
	// (cart.Cart.CanCheckout)。呼び出し元がerrors.Is(err, ErrCartEmpty)で判別できるよう
	// ergo.NewSentinelで定義する(errors.Newで毎回生成すると呼び出し元が種別判定できない)。
	ErrCartEmpty = ergo.NewSentinel("cart has no items")

	// 注文エラー
	// ErrInsufficientStock は注文確定時にカート内商品の在庫が不足している場合の
	// エラー(order.NewFromCart)。呼び出し元がerrors.Is(err, ErrInsufficientStock)で
	// 判別しerrors.MakeBusinessErrorに変換できるよう、ErrCartEmptyと同様に
	// ergo.NewSentinelで定義する。
	ErrInsufficientStock = ergo.NewSentinel("insufficient stock")

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

// WithInvalidArgumentCode はctxを持たずMake*Errorを呼べない箇所
// (echo.Validatorインターフェースの`Validate(i any) error`のようにctx引数が
// 無いシグネチャ)で、クライアント入力起因のエラーだと分かっている場合に
// ErrCodeInvalidArgument(422)だけを付与するためのヘルパー。ログは出力しない
// (ctxが無くログにRequestIDを乗せられないため)。ErrTypeBussinessではラップ
// しないため、呼び出し元がerrors.Wrap(ctx, err)に渡した時点でIsWrapped(err)は
// falseのままとなり、Wrapが実際にログを出力する(このコードは
// ergo.CodeOf経由でWrap後も保持される)。
//
// 現在の利用例: pkg/validator.Validator.Validate（request-validation.md）。
func WithInvalidArgumentCode(err error) error {
	if err == nil {
		return nil
	}
	return ergo.WithCode(err, ErrCodeInvalidArgument)
}

// Wrap は詳細不明な下位レイヤーのエラー(DBドライバ・リポジトリが返す生のエラー等)
// にスタックトレースを付与して伝搬する。ログはこの関数の内部で一度だけ出力する
// (error-handling.md「ラップのみ: 詳細不明の外部エラーはerrors.Wrap(ctx, err)で
// ラップする」)。
//
// 既にMake*Error/Wrap済みのエラー(IsWrapped(err)がtrue)はそのまま返す
// (「一度ラップしたエラーは再ラップしない」ため、二重ログを防ぐ)。
// このWrap自体は新しいエラーコードを付与しないが、呼び出し元が事前に
// WithInvalidArgumentCode等でコードを付与したerrを渡した場合、そのコードは
// ergo.CodeOf経由でErrCodeToStatusAndMessageに引き継がれる。コード未設定の
// ままErrCodeToStatusAndMessageに渡ると500として扱われる。
func Wrap(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if IsWrapped(err) {
		return err
	}
	wrapped := ergo.Wrap(err, err.Error())
	st := ergo.StackTraceOf(wrapped)
	logger.Error(ctx, wrapped.Error(), "callStack", st)
	return wrapped
}

func GetMessage(err error) string {
	return err.Error()
}

func GetCode(err error) ergo.Code {
	return ergo.CodeOf(err)
}

// MakeAuthorizationError は認可失敗(権限不足。HTTP 403相当)のエラーを生成する。
func MakeAuthorizationError(ctx context.Context, msg string) error {
	err := ergo.Wrap(ErrUnauthorized, msg)
	err = ergo.WithCode(err, ErrCodeUnAuthorization)
	st := ergo.StackTraceOf(err)
	logger.Warn(ctx, err.Error(), "callStack", st)
	return err
}

// MakeAuthorizedError は認証失敗(トークン無効など。HTTP 401相当)のエラーを生成する。
func MakeAuthorizedError(ctx context.Context, msg string) error {
	err := ergo.Wrap(ErrAuthorized, msg)
	err = ergo.WithCode(err, ErrCodeUnAuthorized)
	st := ergo.StackTraceOf(err)
	logger.Warn(ctx, err.Error(), "callStack", st)
	return err
}

func MakeBusinessError(ctx context.Context, msg string) error {
	err := ergo.Wrap(ErrTypeBussiness, msg)
	err = ergo.WithCode(err, ErrCodeInvalidArgument)
	st := ergo.StackTraceOf(err)
	logger.Info(ctx, err.Error(), "callStack", st)
	return err
}

func MakeConflictError(ctx context.Context, msg string) error {
	err := ergo.Wrap(ErrTypeConflict, msg)
	err = ergo.WithCode(err, ErrCodeConflict)
	st := ergo.StackTraceOf(err)
	logger.Warn(ctx, err.Error(), "callStack", st)
	return err
}

func MakeNotFoundError(ctx context.Context, msg string) error {
	err := ergo.Wrap(ErrTypeNotFound, msg)
	err = ergo.WithCode(err, ErrCodeNotFound)
	st := ergo.StackTraceOf(err)
	logger.Warn(ctx, err.Error(), "callStack", st)
	return err
}

func MakeSystemError(ctx context.Context, err error) error {
	if !IsWrapped(err) {
		err = ergo.Wrap(ErrTypeCritical, err.Error())
	}
	err = ergo.WithCode(err, ErrCodeSystem)
	st := ergo.StackTraceOf(err)
	logger.Error(ctx, err.Error(), "callStack", st)
	return err
}
