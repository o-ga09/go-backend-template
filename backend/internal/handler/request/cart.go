package request

// CartItemRequest はカートへの商品追加・カート内商品の数量変更リクエストの
// ボディ(フィールドが完全に同一のため1型で共用する)。
type CartItemRequest struct {
	ProductID string `json:"productId" validate:"required" ja:"productIdは必須です"`
	Quantity  int    `json:"quantity" validate:"required,min=1" ja:"quantityは1以上である必要があります"`
}

// RemoveCartItemRequest はカートからの商品削除リクエストのクエリパラメータ。
type RemoveCartItemRequest struct {
	ProductID string `query:"productId" validate:"required" ja:"productIdは必須です"`
}
