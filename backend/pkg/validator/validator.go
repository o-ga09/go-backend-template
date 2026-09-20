// Package validator はecho.Echo#Validatorに登録する、go-playground/validator/v10ベースの
// バリデータを提供する。ハンドラはc.Bindでリクエスト型に値を詰めた後、c.Validateで
// structタグ(validate:"...")によるフォーマット検証を行う。
package validator

import "github.com/go-playground/validator/v10"

// Validator はecho.Validatorインターフェースの実装。
type Validator struct {
	validate *validator.Validate
}

// New はValidatorを生成する。
func New() *Validator {
	return &Validator{validate: validator.New()}
}

// Validate はvalidate:"..."タグに従ってiを検証する。
func (v *Validator) Validate(i any) error {
	return v.validate.Struct(i)
}
