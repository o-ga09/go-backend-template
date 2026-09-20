package request

// GetOrderRequest は注文詳細取得リクエストのパスパラメータ。
type GetOrderRequest struct {
	ID string `param:"id" validate:"required" ja:"idは必須です"`
}
