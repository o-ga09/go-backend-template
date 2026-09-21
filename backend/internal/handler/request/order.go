package request

type GetOrderRequest struct {
	ID string `param:"id" validate:"required" ja:"idは必須です"`
}
