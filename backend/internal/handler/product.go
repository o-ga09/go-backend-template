package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/o-ga09/go-backend-template/internal/domain/product"
	"github.com/o-ga09/go-backend-template/internal/handler/request"
	"github.com/o-ga09/go-backend-template/internal/handler/response"
	"github.com/o-ga09/go-backend-template/pkg/errors"
)

// 商品一覧・詳細は認証不要の公開APIのため、認証チェックは行わない。
type productHandler struct {
	repo product.IProductRepository
}

type IProduct interface {
	List(c *echo.Context) error
	GetByID(c *echo.Context) error
}

func NewProductHandler(repo product.IProductRepository) IProduct {
	return &productHandler{repo: repo}
}

// GET /api/products
func (h *productHandler) List(c *echo.Context) error {
	ctx := c.Request().Context()

	ps, err := h.repo.List(ctx)
	if err != nil {
		return errors.Wrap(ctx, err)
	}

	return c.JSON(http.StatusOK, response.FromProducts(ps))
}

// GET /api/products/:id
func (h *productHandler) GetByID(c *echo.Context) error {
	ctx := c.Request().Context()

	var req request.GetProductRequest
	if err := c.Bind(&req); err != nil {
		return errors.Wrap(ctx, err)
	}
	if err := c.Validate(&req); err != nil {
		return errors.Wrap(ctx, err)
	}

	p, err := h.repo.FindByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
			return errors.MakeNotFoundError(ctx, "product not found")
		}
		return errors.Wrap(ctx, err)
	}

	return c.JSON(http.StatusOK, response.FromProduct(p))
}
