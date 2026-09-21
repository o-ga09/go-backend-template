package request

type GetProductRequest struct {
	ID string `param:"id" validate:"required" ja:"idは必須です"`
}
