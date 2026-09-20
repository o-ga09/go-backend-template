// Package product は商品ドメイン(エンティティ + リポジトリinterface)を定義する。
// 商品一覧・詳細取得は認証不要の公開APIとして提供される
// (backend/tmp/task-2-brief.md参照)。
package product

import (
	"context"

	"github.com/o-ga09/go-backend-template/pkg/model"
)

//go:generate go run github.com/matryer/moq@latest -out mock/product_repository_mock.go -pkg moq . IProductRepository

// Product は商品ドメインエンティティ。domain＝DBモデルの方針に従い、
// GORMモデルを兼ねる(internal/database/mysql.productRepository経由で永続化される)。
type Product struct {
	model.BaseModel
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
	PriceYen    int    `gorm:"column:price_yen"`
	Stock       int    `gorm:"column:stock"`
}

// TableName はGORMが使用するテーブル名を明示する。
func (Product) TableName() string {
	return "products"
}

// IProductRepository は商品ドメインの永続化用インターフェース。
// 実装はinternal/database/mysql.productRepository(GORM)。
// 商品の作成/更新は本Issue(#4)のスコープ外(読み取り専用)。
type IProductRepository interface {
	// FindByID はIDで商品を検索する。見つからない場合はerrors.ErrRecordNotFoundを返す。
	FindByID(ctx context.Context, id string) (*Product, error)
	// List は商品を全件取得する。ページングは今回不要(YAGNI)。
	List(ctx context.Context) ([]*Product, error)
}
