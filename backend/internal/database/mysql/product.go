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

// DecreaseStock は指定数量だけ在庫を減算する。「WHERE stock >= quantity」を
// 条件に含む1回のUPDATEで実装しており、これ自体が在庫不足の判定を兼ねる
// (RowsAffected == 0なら在庫不足)。事前チェック(HasStock)からこの呼び出しまでの
// 間に他のリクエストが在庫を消費しても、条件付きUPDATEにより在庫がマイナスに
// なることはない(TOCTOU耐性)。
//
// UpdateColumnはBaseModelPluginのUpdateフック(楽観ロック)の対象外になる
// (呼び出しに使うproduct.Product{}のVersionがゼロ値のため、
// base_model_plugin.goのbeforeUpdateが早期returnする)。在庫の排他制御は
// この条件付きUPDATE自体が担うため、Versionによる楽観ロックとは独立している。
func (r *productRepository) DecreaseStock(ctx context.Context, productID string, quantity int) error {
	db := Ctx.GetDBFromCtx(ctx)

	result := db.WithContext(ctx).Model(&product.Product{}).
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
