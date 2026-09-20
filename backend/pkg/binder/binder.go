// Package binder はecho.Echo#Binderに登録するカスタムバインダを提供する。
//
// このリポジトリが使うecho v3(labstack/echo v3.3.10)のDefaultBinderは
// パスパラメータ(`param`タグ)のバインディングをサポートしていない(echo v4には
// あるがv3にはない)。ハンドラのリクエスト型でタグを一貫して使い分けられるよう
// (param→パスパラメータ、query→クエリパラメータ、json→リクエストボディ)、
// パスパラメータのバインディングだけを補うラッパーとして実装する。
package binder

import (
	"reflect"
	"strings"

	"github.com/labstack/echo"
)

type pathParamBinder struct {
	fallback echo.Binder
}

// New はDefaultBinderにパスパラメータ(`param`タグ)バインディングを追加したechoBinderを生成する。
func New() echo.Binder {
	return &pathParamBinder{fallback: &echo.DefaultBinder{}}
}

// Bind はまず`param`タグのフィールドをc.Param()から埋め、続けてecho標準のDefaultBinderで
// クエリパラメータ・リクエストボディをバインドする。
func (b *pathParamBinder) Bind(i any, c echo.Context) error {
	bindPathParams(i, c)
	return b.fallback.Bind(i, c)
}

func bindPathParams(i any, c echo.Context) {
	val := reflect.ValueOf(i)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return
	}
	val = val.Elem()
	typ := val.Type()
	for idx := range typ.NumField() {
		field := typ.Field(idx)
		tag, ok := field.Tag.Lookup("param")
		if !ok {
			continue
		}
		name := strings.Split(tag, ",")[0]
		value := c.Param(name)
		structField := val.Field(idx)
		if value == "" || !structField.CanSet() || structField.Kind() != reflect.String {
			continue
		}
		structField.SetString(value)
	}
}
