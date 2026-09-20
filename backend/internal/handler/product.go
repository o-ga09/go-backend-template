package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/o-ga09/go-backend-template/internal/domain/product"
	"github.com/o-ga09/go-backend-template/internal/handler/request"
	"github.com/o-ga09/go-backend-template/internal/handler/response"
	"github.com/o-ga09/go-backend-template/pkg/errors"
)

// productHandler は商品リソース(/api/products)に関するエンドポイントを扱う。
// 一覧・詳細取得のみのシンプルなCRUDのためusecase層を挟まず、domainの
// リポジトリを直接呼び出す(architecture.md「基本方針：レイヤードアーキテクチャ」)。
// 商品一覧・詳細は認証不要の公開APIのため、認証チェックは行わない
// (backend/tmp/task-2-brief.md参照)。
type productHandler struct {
	repo product.IProductRepository
}

// IProduct はproductHandlerの公開インターフェース。
type IProduct interface {
	List(c *echo.Context) error
	GetByID(c *echo.Context) error
}

// NewProductHandler はproductHandlerを生成する。
func NewProductHandler(repo product.IProductRepository) IProduct {
	return &productHandler{repo: repo}
}

// List は商品一覧を取得する。
// GET /api/products
func (h *productHandler) List(c *echo.Context) error {
	ctx := c.Request().Context()

	ps, err := h.repo.List(ctx)
	if err != nil {
		return errors.Wrap(ctx, err)
	}

	return c.JSON(http.StatusOK, response.FromProducts(ps))
}

// GetByID は商品詳細を取得する。存在しない場合は404を返す。
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
