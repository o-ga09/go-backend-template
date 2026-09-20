package request

// AddCartItemRequest はカートへの商品追加リクエストのボディ。
type AddCartItemRequest struct {
	ProductID string `json:"productId" validate:"required" ja:"productIdは必須です"`
	Quantity  int    `json:"quantity" validate:"required,min=1" ja:"quantityは1以上である必要があります"`
}

// UpdateCartItemRequest はカート内商品の数量変更リクエストのボディ。
type UpdateCartItemRequest struct {
	ProductID string `json:"productId" validate:"required" ja:"productIdは必須です"`
	Quantity  int    `json:"quantity" validate:"required,min=1" ja:"quantityは1以上である必要があります"`
}

// RemoveCartItemRequest はカートからの商品削除リクエストのクエリパラメータ。
type RemoveCartItemRequest struct {
	ProductID string `query:"productId" validate:"required" ja:"productIdは必須です"`
}
