package mysql

import (
	"context"
	stderrors "errors"

	"gorm.io/gorm"

	"github.com/o-ga09/go-backend-template/internal/domain/product"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	pkgerrors "github.com/o-ga09/go-backend-template/pkg/errors"
)

// productRepository はproductドメインのGORMリポジトリ。
// *gorm.DBをフィールドに保持せず、呼び出しごとにctxから取得する(user.goと同じ理由)。
// 商品の作成/更新は本Issue(#4)のスコープ外(読み取り専用)。
type productRepository struct{}

// NewProductRepository はproductRepositoryを生成する。
func NewProductRepository() product.IProductRepository {
	return &productRepository{}
}

// FindByID はIDで商品を検索する。
func (r *productRepository) FindByID(ctx context.Context, id string) (*product.Product, error) {
	db := Ctx.GetDBFromCtx(ctx)

	var p product.Product
	if err := db.WithContext(ctx).Where("id = ?", id).First(&p).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &p, nil
}

// List は商品を全件取得する。ページングは今回不要(YAGNI)。
// 返却順は作成日時の降順で安定させる(ORDER BY未指定だと順序が不定になるため)。
func (r *productRepository) List(ctx context.Context) ([]*product.Product, error) {
	db := Ctx.GetDBFromCtx(ctx)

	var ps []*product.Product
	if err := db.WithContext(ctx).Order("created_at DESC").Find(&ps).Error; err != nil {
		return nil, err
	}
	return ps, nil
}
