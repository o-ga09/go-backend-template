package validator_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/o-ga09/go-backend-template/pkg/validator"
)

func TestValidator_Validate(t *testing.T) {
	type request struct {
		UID   string `validate:"required" ja:"uidは必須です"`
		Email string `validate:"omitempty,email" ja:"emailの形式が正しくありません"`
		Age   int    `validate:"gte=0"`
	}

	tests := []struct {
		name    string
		req     request
		wantErr string
	}{
		{"全フィールドが妥当な場合はエラー無し", request{UID: "u1", Email: "a@example.com", Age: 20}, ""},
		{"必須フィールドが空の場合はjaタグのメッセージを返す", request{UID: "", Age: 20}, "uidは必須です"},
		{"フォーマット違反の場合もjaタグのメッセージを返す", request{UID: "u1", Email: "invalid", Age: 20}, "emailの形式が正しくありません"},
		{"jaタグが無いフィールドはデフォルトメッセージにフォールバックする", request{UID: "u1", Age: -1}, "Age"},
		{"複数フィールドが違反した場合はメッセージを連結する", request{UID: "", Email: "invalid", Age: 20}, "uidは必須です、emailの形式が正しくありません"},
	}

	v := validator.New()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Validate(&tc.req)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
			// Validate()はerrors.WithInvalidArgumentCode(422)で包んで返す
			// (request-validation.md)ため、jaタグメッセージ自体は
			// errors.Asで*validator.ValidationErrorまで辿って検証する。
			var verr *validator.ValidationError
			if !errors.As(err, &verr) {
				t.Fatalf("Validate() error chain does not contain *validator.ValidationError: %v", err)
			}
			if got := verr.Error(); got != tc.wantErr {
				// デフォルトメッセージへのフォールバックは完全一致ではなくフィールド名を含むかで確認する
				if tc.name == "jaタグが無いフィールドはデフォルトメッセージにフォールバックする" {
					if !strings.Contains(got, tc.wantErr) {
						t.Fatalf("Validate() error = %q, want it to contain %q", got, tc.wantErr)
					}
					return
				}
				t.Fatalf("Validate() error = %q, want %q", got, tc.wantErr)
			}
		})
	}
}
