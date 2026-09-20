// Package validator はecho.Echo#Validatorに登録する、go-playground/validator/v10ベースの
// バリデータを提供する。ハンドラはc.Bindでリクエスト型に値を詰めた後、c.Validateで
// structタグ(validate:"...")によるフォーマット検証を行う。
//
// バリデーション失敗時にクライアントへ返すメッセージは、structタグ`ja`で
// フィールドごとに指定する(例: `validate:"required" ja:"uidは必須です"`)。
// `ja`タグが無いフィールドはgo-playground/validatorのデフォルト(英語)メッセージに
// フォールバックする。
package validator

import (
	"reflect"
	"strings"

	val "github.com/go-playground/validator/v10"
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
// `ja`タグを使って組み立てたエラーを返す。
func (v *Validator) Validate(i any) error {
	err := v.validate.Struct(i)
	if err == nil {
		return nil
	}

	verrs, ok := err.(val.ValidationErrors)
	if !ok {
		return err
	}
	return translate(i, verrs)
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
		if field, ok := typ.FieldByName(fe.StructField()); ok {
			if ja, ok := field.Tag.Lookup("ja"); ok && ja != "" {
				return ja
			}
		}
	}
	return fe.Error()
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
