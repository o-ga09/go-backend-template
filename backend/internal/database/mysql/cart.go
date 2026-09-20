package mysql

import (
	"context"
	stderrors "errors"

	"gorm.io/gorm"

	"github.com/o-ga09/go-backend-template/internal/domain/cart"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	pkgerrors "github.com/o-ga09/go-backend-template/pkg/errors"
)

// cartRepository はcartドメインのGORMリポジトリ。
// *gorm.DBをフィールドに保持せず、呼び出しごとにctxから取得する(user.goと同じ理由)。
// リポジトリはデータの読み書きのみを行い、ビジネスロジックを持たない
// (商品の存在確認・在庫確認はハンドラ層で行う。architecture.mdの依存方向を厳守)。
type cartRepository struct{}

// NewCartRepository はcartRepositoryを生成する。
func NewCartRepository() cart.ICartRepository {
	return &cartRepository{}
}

// FindByUserID はユーザーIDでカート(明細含む)を検索する。
func (r *cartRepository) FindByUserID(ctx context.Context, userID string) (*cart.Cart, error) {
	db := Ctx.GetDBFromCtx(ctx)

	var c cart.Cart
	if err := db.WithContext(ctx).Where("user_id = ?", userID).First(&c).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrRecordNotFound
		}
		return nil, err
	}

	items, err := r.findItems(ctx, db, c.ID)
	if err != nil {
		return nil, err
	}
	c.Items = items
	return &c, nil
}

// findItems はカートIDに紐づく明細を作成日時の昇順で取得する。
func (r *cartRepository) findItems(ctx context.Context, db *gorm.DB, cartID string) ([]cart.CartItem, error) {
	var items []cart.CartItem
	if err := db.WithContext(ctx).Where("cart_id = ?", cartID).Order("created_at ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// Create はユーザーのカートを新規作成する。ID/Version/タイムスタンプはBaseModelPluginが採番する。
func (r *cartRepository) Create(ctx context.Context, userID string) (*cart.Cart, error) {
	db := Ctx.GetDBFromCtx(ctx)

	c := &cart.Cart{UserID: userID}
	if err := db.WithContext(ctx).Create(c).Error; err != nil {
		if isDuplicateEntryErr(err) {
			return nil, pkgerrors.ErrUniqueConstraint
		}
		return nil, err
	}
	return c, nil
}

// AddItem はカートに商品を追加する。既に同じ商品が入っている場合は数量を加算する
// (cart_items.uq_cart_items_cart_productのUNIQUE制約に対応するため)。
func (r *cartRepository) AddItem(ctx context.Context, cartID, productID string, quantity int) error {
	db := Ctx.GetDBFromCtx(ctx)

	var item cart.CartItem
	err := db.WithContext(ctx).Where("cart_id = ? AND product_id = ?", cartID, productID).First(&item).Error
	if err == nil {
		return db.WithContext(ctx).Model(&item).Updates(map[string]interface{}{
			"quantity": item.Quantity + quantity,
		}).Error
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	newItem := &cart.CartItem{CartID: cartID, ProductID: productID, Quantity: quantity}
	if err := db.WithContext(ctx).Create(newItem).Error; err != nil {
		if isDuplicateEntryErr(err) {
			return pkgerrors.ErrUniqueConstraint
		}
		return err
	}
	return nil
}

// UpdateItemQuantity はカート内の商品の数量を指定した値に変更する。
func (r *cartRepository) UpdateItemQuantity(ctx context.Context, cartID, productID string, quantity int) error {
	db := Ctx.GetDBFromCtx(ctx)

	var item cart.CartItem
	if err := db.WithContext(ctx).Where("cart_id = ? AND product_id = ?", cartID, productID).First(&item).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return pkgerrors.ErrRecordNotFound
		}
		return err
	}

	return db.WithContext(ctx).Model(&item).Updates(map[string]interface{}{
		"quantity": quantity,
	}).Error
}

// RemoveItem はカートから商品を削除する。
func (r *cartRepository) RemoveItem(ctx context.Context, cartID, productID string) error {
	db := Ctx.GetDBFromCtx(ctx)

	res := db.WithContext(ctx).Where("cart_id = ? AND product_id = ?", cartID, productID).Delete(&cart.CartItem{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return pkgerrors.ErrRecordNotFound
	}
	return nil
}

// Clear はカート内の全明細を削除する(注文確定時にTask 4から呼ばれる想定)。
func (r *cartRepository) Clear(ctx context.Context, cartID string) error {
	db := Ctx.GetDBFromCtx(ctx)

	return db.WithContext(ctx).Where("cart_id = ?", cartID).Delete(&cart.CartItem{}).Error
}
