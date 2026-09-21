package mysql

import (
	"context"
	stderrors "errors"

	"gorm.io/gorm"

	"github.com/o-ga09/go-backend-template/internal/domain/cart"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	pkgerrors "github.com/o-ga09/go-backend-template/pkg/errors"
)

// リポジトリはデータの読み書きのみを行う(商品の存在確認・在庫確認はハンドラ層で行う。
// architecture.mdの依存方向を厳守)。
type cartRepository struct{}

func NewCartRepository() cart.ICartRepository {
	return &cartRepository{}
}

func (r *cartRepository) FindByUserID(ctx context.Context, userID string) (*cart.Cart, error) {
	var c cart.Cart
	err := Ctx.GetDBFromCtx(ctx).
		Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Order("created_at ASC, id ASC") }).
		Where("user_id = ?", userID).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *cartRepository) Create(ctx context.Context, userID string) (*cart.Cart, error) {
	c := &cart.Cart{UserID: userID}
	if err := Ctx.GetDBFromCtx(ctx).Create(c).Error; err != nil {
		if isDuplicateEntryErr(err) {
			return nil, pkgerrors.ErrUniqueConstraint
		}
		return nil, err
	}
	return c, nil
}

// 既に同じ商品が入っている場合は数量を加算する(cart_items.uq_cart_items_cart_product
// のUNIQUE制約に対応するため)。
func (r *cartRepository) AddItem(ctx context.Context, cartID, productID string, quantity int) error {
	db := Ctx.GetDBFromCtx(ctx)

	var item cart.CartItem
	err := db.Where("cart_id = ? AND product_id = ?", cartID, productID).First(&item).Error
	if err == nil {
		return db.Model(&item).Updates(map[string]interface{}{
			"quantity": item.Quantity + quantity,
		}).Error
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	newItem := &cart.CartItem{CartID: cartID, ProductID: productID, Quantity: quantity}
	if err := db.Create(newItem).Error; err != nil {
		if isDuplicateEntryErr(err) {
			return pkgerrors.ErrUniqueConstraint
		}
		return err
	}
	return nil
}

func (r *cartRepository) UpdateItemQuantity(ctx context.Context, cartID, productID string, quantity int) error {
	db := Ctx.GetDBFromCtx(ctx)

	var item cart.CartItem
	if err := db.Where("cart_id = ? AND product_id = ?", cartID, productID).First(&item).Error; err != nil {
		return err
	}

	return db.Model(&item).Updates(map[string]interface{}{
		"quantity": quantity,
	}).Error
}

func (r *cartRepository) RemoveItem(ctx context.Context, cartID, productID string) error {
	res := Ctx.GetDBFromCtx(ctx).Where("cart_id = ? AND product_id = ?", cartID, productID).Delete(&cart.CartItem{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return pkgerrors.ErrRecordNotFound
	}
	return nil
}

func (r *cartRepository) Clear(ctx context.Context, cartID string) error {
	return Ctx.GetDBFromCtx(ctx).Where("cart_id = ?", cartID).Delete(&cart.CartItem{}).Error
}
