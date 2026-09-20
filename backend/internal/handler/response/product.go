package response

import "github.com/o-ga09/go-backend-template/internal/domain/product"

// Product はクライアントに返す商品情報。
type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceYen    int    `json:"priceYen"`
	Stock       int    `json:"stock"`
}

// FromProduct はdomain.ProductからレスポンスのProductを組み立てる。
func FromProduct(p *product.Product) Product {
	return Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		PriceYen:    p.PriceYen,
		Stock:       p.Stock,
	}
}

// FromProducts はdomain.Productのスライスからレスポンスのスライスを組み立てる。
func FromProducts(ps []*product.Product) []Product {
	res := make([]Product, 0, len(ps))
	for _, p := range ps {
		res = append(res, FromProduct(p))
	}
	return res
}
