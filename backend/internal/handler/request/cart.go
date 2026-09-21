package request

// 商品追加と数量変更でフィールドが完全に同一のため1型で共用する。
type CartItemRequest struct {
	ProductID string `json:"productId" validate:"required" ja:"productIdは必須です"`
	Quantity  int    `json:"quantity" validate:"required,min=1" ja:"quantityは1以上である必要があります"`
}

type RemoveCartItemRequest struct {
	ProductID string `query:"productId" validate:"required" ja:"productIdは必須です"`
}
