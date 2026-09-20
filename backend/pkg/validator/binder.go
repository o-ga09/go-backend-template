package validator

import (
	"github.com/labstack/echo/v5"

	"github.com/o-ga09/go-backend-template/pkg/errors"
)

// Binder はecho.Binderインターフェースの実装。バインド自体はecho.DefaultBinderに
// 委譲し、失敗した場合のみerrors.WithInvalidArgumentCode(422)を付与して返す。
// echo.Binderインターフェースの`Bind(c *echo.Context, target any) error`にはctxが
// 渡らないためerrors.MakeBusinessErrorは呼べないが、Validateと同様にコードだけ
// 先に付けておけば、ctxを持つハンドラ側でerrors.Wrap(ctx, err)するだけで422が
// 自動的にレスポンスされる(request-validation.md参照)。
type Binder struct {
	delegate echo.Binder
}

// NewBinder はBinderを生成する。
func NewBinder() *Binder {
	return &Binder{delegate: &echo.DefaultBinder{}}
}

func (b *Binder) Bind(c *echo.Context, target any) error {
	if err := b.delegate.Bind(c, target); err != nil {
		return errors.WithInvalidArgumentCode(err)
	}
	return nil
}
