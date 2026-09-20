// Package validator はecho.Echo#Validatorに登録する、go-playground/validator/v10ベースの
// バリデータを提供する。ハンドラはc.Bindでリクエスト型に値を詰めた後、c.Validateで
// structタグ(validate:"...")によるフォーマット検証を行う。
//
// バリデーション失敗時にクライアントへ返すメッセージは、structタグ`ja`で
// フィールドごとに指定する(例: `validate:"required" ja:"uidは必須です"`)。
// `ja`タグが無いフィールドはgo-playground/validatorのデフォルト(英語)メッセージに
// フォールバックする。
//
// Validateが返すエラーには、その場でerrors.WithInvalidArgumentCode(422)を
// 付与しておく。echo.Validatorインターフェースの`Validate(i any) error`には
// ctxが渡らないためerrors.MakeBusinessErrorは呼べないが、コードだけ先に
// 付けておけば、ctxを持つハンドラ側でerrors.Wrap(ctx, err)するだけで422が
// 自動的にレスポンスされる(request-validation.md参照)。
package validator

import (
	"reflect"
	"strings"

	val "github.com/go-playground/validator/v10"

	"github.com/o-ga09/go-backend-template/pkg/errors"
)

// Validator はecho.Validatorインターフェースの実装。
type Validator struct {
	validate *val.Validate
}

// New はValidatorを生成する。
func New() *Validator {
	return &Validator{validate: val.New()}
}

// Validate はvalidate:"..."タグに従ってiを検証する。失敗した場合、失敗した各フィールドの
// `ja`タグを使って組み立てたエラーに422のエラーコードを付与して返す
// (errors.WithInvalidArgumentCode。呼び出し元はerrors.Wrap(ctx, err)するだけでよい)。
func (v *Validator) Validate(i any) error {
	err := v.validate.Struct(i)
	if err == nil {
		return nil
	}

	verrs, ok := err.(val.ValidationErrors)
	if !ok {
		return err
	}
	return errors.WithInvalidArgumentCode(translate(i, verrs))
}

// translate はvalidator.ValidationErrorsを、各フィールドの`ja`タグを使った
// 日本語メッセージのValidationErrorに変換する。
func translate(i any, verrs val.ValidationErrors) error {
	typ := reflect.TypeOf(i)
	for typ != nil && typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	messages := make([]string, 0, len(verrs))
	for _, fe := range verrs {
		messages = append(messages, messageFor(typ, fe))
	}
	return &ValidationError{Messages: messages}
}

// messageFor は1件のフィールドエラーに対応するメッセージを返す。
// 対象フィールドに`ja`タグがあればその値を、無ければgo-playground/validatorの
// デフォルトメッセージ(fe.Error())を返す。
func messageFor(typ reflect.Type, fe val.FieldError) string {
	if typ != nil && typ.Kind() == reflect.Struct {
		if ja, ok := jaTagFor(typ, fe.StructField()); ok && ja != "" {
			return ja
		}
	}
	return fe.Error()
}

// jaTagFor はtypの各フィールドを順に走査し、名前がfieldNameと一致するフィールドの
// `ja`タグを返す。fieldNameはgo-playground/validatorがtyp自身の構造体定義から
// 決定した値でリクエスト由来ではないが、reflect.Type.FieldByNameのような
// 「名前を動的に指定してフィールド/メソッドを解決する」API(静的解析で
// go.lang.security.audit.unsafe-reflect-by-nameとして検出されやすい)は使わず、
// 明示的なループで比較する。
func jaTagFor(typ reflect.Type, fieldName string) (string, bool) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Name != fieldName {
			continue
		}
		ja, ok := field.Tag.Lookup("ja")
		return ja, ok
	}
	return "", false
}

// ValidationError はバリデーション失敗時のエラー。失敗した各フィールドの
// メッセージ(`ja`タグ、無ければデフォルトメッセージ)を保持する。
type ValidationError struct {
	Messages []string
}

// Error はメッセージを「、」区切りで連結して返す。
func (e *ValidationError) Error() string {
	return strings.Join(e.Messages, "、")
}
