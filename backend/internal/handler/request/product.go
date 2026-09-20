package request

// GetProductRequest は商品詳細取得リクエストのパスパラメータ。
type GetProductRequest struct {
	ID string `param:"id" validate:"required" ja:"idは必須です"`
}
