package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/o-ga09/go-backend-template/internal/domain/product"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	pkgerrors "github.com/o-ga09/go-backend-template/pkg/errors"
)

// 商品の作成/更新は本Issue(#4)のスコープ外(読み取り専用。DecreaseStockのみ例外)。
type productRepository struct{}

func NewProductRepository() product.IProductRepository {
	return &productRepository{}
}

func (r *productRepository) FindByID(ctx context.Context, id string) (*product.Product, error) {
	var p product.Product
	if err := Ctx.GetDBFromCtx(ctx).Where("id = ?", id).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// ORDER BY未指定だと返却順が不定になるため作成日時の降順を明示する。
func (r *productRepository) List(ctx context.Context) ([]*product.Product, error) {
	var ps []*product.Product
	if err := Ctx.GetDBFromCtx(ctx).Order("created_at DESC").Find(&ps).Error; err != nil {
		return nil, err
	}
	return ps, nil
}

// 条件付きUPDATE(stock >= quantity)自体が在庫不足の判定を兼ねるため、事前チェック
// (HasStock)からこの呼び出しまでの間の競合(TOCTOU)でも在庫はマイナスにならない。
// Modelに渡すproduct.Product{}はVersionがゼロ値のため、BaseModelPluginの楽観ロック
// フック(beforeUpdate)は早期returnし介入しない。
func (r *productRepository) DecreaseStock(ctx context.Context, productID string, quantity int) error {
	result := Ctx.GetDBFromCtx(ctx).Model(&product.Product{}).
		Where("id = ? AND stock >= ?", productID, quantity).
		UpdateColumn("stock", gorm.Expr("stock - ?", quantity))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return pkgerrors.ErrInsufficientStock
	}
	return nil
}
