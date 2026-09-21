package cart

import (
	"context"

	"github.com/o-ga09/go-backend-template/pkg/model"
)

//go:generate go run github.com/matryer/moq@latest -out mock/cart_repository_mock.go -pkg moq . ICartRepository

// 1ユーザー1カート(carts.user_idにUNIQUE制約)。
type Cart struct {
	model.BaseModel
	UserID string     `gorm:"column:user_id"`
	Items  []CartItem `gorm:"foreignKey:CartID"`
}

func (Cart) TableName() string {
	return "carts"
}

type CartItem struct {
	model.BaseModel
	CartID    string `gorm:"column:cart_id"`
	ProductID string `gorm:"column:product_id"`
	Quantity  int    `gorm:"column:quantity"`
}

func (CartItem) TableName() string {
	return "cart_items"
}

type ICartRepository interface {
	FindByUserID(ctx context.Context, userID string) (*Cart, error)
	Create(ctx context.Context, userID string) (*Cart, error)
	AddItem(ctx context.Context, cartID, productID string, quantity int) error
	UpdateItemQuantity(ctx context.Context, cartID, productID string, quantity int) error
	RemoveItem(ctx context.Context, cartID, productID string) error
	Clear(ctx context.Context, cartID string) error
}
