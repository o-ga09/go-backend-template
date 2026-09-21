package product

import (
	"context"

	"github.com/o-ga09/go-backend-template/pkg/model"
)

//go:generate go run github.com/matryer/moq@latest -out mock/product_repository_mock.go -pkg moq . IProductRepository

type Product struct {
	model.BaseModel
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
	PriceYen    int    `gorm:"column:price_yen"`
	Stock       int    `gorm:"column:stock"`
}

func (Product) TableName() string {
	return "products"
}

// 商品の作成/更新は本Issue(#4)のスコープ外だが、注文確定に伴う在庫減算
// (DecreaseStock)のみ例外として持つ。
type IProductRepository interface {
	FindByID(ctx context.Context, id string) (*Product, error)
	List(ctx context.Context) ([]*Product, error)
	// 条件付きUPDATE(stock >= quantity)で実装するため、事前チェック(HasStock)との
	// 間の競合(TOCTOU)でも在庫はマイナスにならない。在庫不足時はerrors.ErrInsufficientStock。
	DecreaseStock(ctx context.Context, productID string, quantity int) error
}
