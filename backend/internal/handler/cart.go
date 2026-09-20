package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/o-ga09/go-backend-template/internal/domain/cart"
	"github.com/o-ga09/go-backend-template/internal/domain/product"
	"github.com/o-ga09/go-backend-template/internal/handler/request"
	"github.com/o-ga09/go-backend-template/internal/handler/response"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	"github.com/o-ga09/go-backend-template/pkg/errors"
)

// cartHandler はカートリソース(/api/cart)に関するエンドポイントを扱う。
// シンプルなCRUDのためusecase層を挟まず、domainのリポジトリを直接呼び出す
// (architecture.md「基本方針：レイヤードアーキテクチャ」)。
// カート操作は全て認証必須で、常にログイン中ユーザー本人のカートのみを対象とする
// (カートIDをリクエストから受け取らないため、所有者チェックは不要。
// backend/tmp/task-3-brief.md参照)。
type cartHandler struct {
	cartRepo    cart.ICartRepository
	productRepo product.IProductRepository
}

// ICart はcartHandlerの公開インターフェース。
type ICart interface {
	Get(c *echo.Context) error
	AddItem(c *echo.Context) error
	UpdateItem(c *echo.Context) error
	RemoveItem(c *echo.Context) error
}

// NewCartHandler はcartHandlerを生成する。
func NewCartHandler(cartRepo cart.ICartRepository, productRepo product.IProductRepository) ICart {
	return &cartHandler{cartRepo: cartRepo, productRepo: productRepo}
}

// Get はログイン中ユーザーのカートを取得する。カートが未作成の場合は
// 空カート相当のレスポンスを返す(404にはしない)。
// GET /api/cart
func (h *cartHandler) Get(c *echo.Context) error {
	ctx := c.Request().Context()

	requesterID := Ctx.GetUserID(ctx)
	if requesterID == "" {
		return errors.MakeAuthorizedError(ctx, "authentication required")
	}

	ct, err := h.cartRepo.FindByUserID(ctx, requesterID)
	if err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
			return c.JSON(http.StatusOK, response.FromCart(&cart.Cart{}))
		}
		return errors.Wrap(ctx, err)
	}

	return c.JSON(http.StatusOK, response.FromCart(ct))
}

// AddItem はログイン中ユーザーのカートに商品を追加する。商品が存在しない場合は404、
// 在庫が不足している場合は422を返す。カートが未作成の場合は新規作成してから追加する。
// POST /api/cart
func (h *cartHandler) AddItem(c *echo.Context) error {
	ctx := c.Request().Context()

	requesterID := Ctx.GetUserID(ctx)
	if requesterID == "" {
		return errors.MakeAuthorizedError(ctx, "authentication required")
	}

	var req request.AddCartItemRequest
	if err := c.Bind(&req); err != nil {
		return errors.Wrap(ctx, err)
	}
	if err := c.Validate(&req); err != nil {
		return errors.Wrap(ctx, err)
	}

	p, err := h.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
			return errors.MakeNotFoundError(ctx, "product not found")
		}
		return errors.Wrap(ctx, err)
	}
	if !p.HasStock(req.Quantity) {
		return errors.MakeBusinessError(ctx, "insufficient stock")
	}

	ct, err := h.cartRepo.FindByUserID(ctx, requesterID)
	if err != nil {
		if !errors.Is(err, errors.ErrRecordNotFound) {
			return errors.Wrap(ctx, err)
		}
		ct, err = h.cartRepo.Create(ctx, requesterID)
		if err != nil {
			return errors.Wrap(ctx, err)
		}
	}

	if err := h.cartRepo.AddItem(ctx, ct.ID, req.ProductID, req.Quantity); err != nil {
		return errors.Wrap(ctx, err)
	}

	updated, err := h.cartRepo.FindByUserID(ctx, requesterID)
	if err != nil {
		return errors.Wrap(ctx, err)
	}
	return c.JSON(http.StatusOK, response.FromCart(updated))
}

// UpdateItem はログイン中ユーザーのカート内商品の数量を変更する。
// カート、または対象の明細が存在しない場合は404を返す。
// PUT /api/cart
func (h *cartHandler) UpdateItem(c *echo.Context) error {
	ctx := c.Request().Context()

	requesterID := Ctx.GetUserID(ctx)
	if requesterID == "" {
		return errors.MakeAuthorizedError(ctx, "authentication required")
	}

	var req request.UpdateCartItemRequest
	if err := c.Bind(&req); err != nil {
		return errors.Wrap(ctx, err)
	}
	if err := c.Validate(&req); err != nil {
		return errors.Wrap(ctx, err)
	}

	ct, err := h.cartRepo.FindByUserID(ctx, requesterID)
	if err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
			return errors.MakeNotFoundError(ctx, "cart not found")
		}
		return errors.Wrap(ctx, err)
	}

	if err := h.cartRepo.UpdateItemQuantity(ctx, ct.ID, req.ProductID, req.Quantity); err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
			return errors.MakeNotFoundError(ctx, "cart item not found")
		}
		return errors.Wrap(ctx, err)
	}

	updated, err := h.cartRepo.FindByUserID(ctx, requesterID)
	if err != nil {
		return errors.Wrap(ctx, err)
	}
	return c.JSON(http.StatusOK, response.FromCart(updated))
}

// RemoveItem はログイン中ユーザーのカートから商品を削除する。
// カート、または対象の明細が存在しない場合は404を返す。
// DELETE /api/cart
func (h *cartHandler) RemoveItem(c *echo.Context) error {
	ctx := c.Request().Context()

	requesterID := Ctx.GetUserID(ctx)
	if requesterID == "" {
		return errors.MakeAuthorizedError(ctx, "authentication required")
	}

	var req request.RemoveCartItemRequest
	if err := c.Bind(&req); err != nil {
		return errors.Wrap(ctx, err)
	}
	if err := c.Validate(&req); err != nil {
		return errors.Wrap(ctx, err)
	}

	ct, err := h.cartRepo.FindByUserID(ctx, requesterID)
	if err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
			return errors.MakeNotFoundError(ctx, "cart not found")
		}
		return errors.Wrap(ctx, err)
	}

	if err := h.cartRepo.RemoveItem(ctx, ct.ID, req.ProductID); err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
			return errors.MakeNotFoundError(ctx, "cart item not found")
		}
		return errors.Wrap(ctx, err)
	}

	updated, err := h.cartRepo.FindByUserID(ctx, requesterID)
	if err != nil {
		return errors.Wrap(ctx, err)
	}
	return c.JSON(http.StatusOK, response.FromCart(updated))
}
