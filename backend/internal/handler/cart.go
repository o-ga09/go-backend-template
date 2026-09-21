package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/o-ga09/go-backend-template/internal/database"
	"github.com/o-ga09/go-backend-template/internal/domain/cart"
	"github.com/o-ga09/go-backend-template/internal/domain/product"
	"github.com/o-ga09/go-backend-template/internal/handler/request"
	"github.com/o-ga09/go-backend-template/internal/handler/response"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	"github.com/o-ga09/go-backend-template/pkg/errors"
)

// カートIDをリクエストから受け取らず常にrequesterIDでスコープするため、
// 所有者チェックは不要。
type cartHandler struct {
	cartRepo    cart.ICartRepository
	productRepo product.IProductRepository
	txManager   database.ITransactionManager
}

type ICart interface {
	Get(c *echo.Context) error
	AddItem(c *echo.Context) error
	UpdateItem(c *echo.Context) error
	RemoveItem(c *echo.Context) error
}

func NewCartHandler(cartRepo cart.ICartRepository, productRepo product.IProductRepository, txManager database.ITransactionManager) ICart {
	return &cartHandler{cartRepo: cartRepo, productRepo: productRepo, txManager: txManager}
}

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

// POST /api/cart
func (h *cartHandler) AddItem(c *echo.Context) error {
	ctx := c.Request().Context()

	requesterID := Ctx.GetUserID(ctx)
	if requesterID == "" {
		return errors.MakeAuthorizedError(ctx, "authentication required")
	}

	var req request.CartItemRequest
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

	// 既存数量を無視すると複数回の追加で在庫超過したカートが作れてしまうため、
	// 加算後の数量で在庫を判定する。
	existing, err := h.cartRepo.FindByUserID(ctx, requesterID)
	cartExists := true
	if err != nil {
		if !errors.Is(err, errors.ErrRecordNotFound) {
			return errors.Wrap(ctx, err)
		}
		cartExists = false
	}

	existingQuantity := 0
	if cartExists {
		existingQuantity = quantityOf(existing, req.ProductID)
	}
	if !p.HasStock(existingQuantity + req.Quantity) {
		return errors.MakeBusinessError(ctx, "insufficient stock")
	}

	cartID := ""
	if cartExists {
		cartID = existing.ID
	}
	if err := h.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		if cartID == "" {
			created, err := h.cartRepo.Create(txCtx, requesterID)
			if err != nil {
				if errors.Is(err, errors.ErrUniqueConstraint) {
					// 同時リクエストで既に作成されていた場合はそちらを使う。
					found, ferr := h.cartRepo.FindByUserID(txCtx, requesterID)
					if ferr != nil {
						return ferr
					}
					created = found
				} else {
					return err
				}
			}
			cartID = created.ID
		}
		return h.cartRepo.AddItem(txCtx, cartID, req.ProductID, req.Quantity)
	}); err != nil {
		if errors.Is(err, errors.ErrUniqueConstraint) || errors.Is(err, errors.ErrOptimisticLockConflict) {
			return errors.MakeConflictError(ctx, "cart update conflict")
		}
		return errors.Wrap(ctx, err)
	}

	updated, err := h.cartRepo.FindByUserID(ctx, requesterID)
	if err != nil {
		return errors.Wrap(ctx, err)
	}
	return c.JSON(http.StatusOK, response.FromCart(updated))
}

// PUT /api/cart
func (h *cartHandler) UpdateItem(c *echo.Context) error {
	ctx := c.Request().Context()

	requesterID := Ctx.GetUserID(ctx)
	if requesterID == "" {
		return errors.MakeAuthorizedError(ctx, "authentication required")
	}

	var req request.CartItemRequest
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
		if errors.Is(err, errors.ErrRecordNotFound) {
			return errors.MakeNotFoundError(ctx, "cart not found")
		}
		return errors.Wrap(ctx, err)
	}

	if err := h.cartRepo.UpdateItemQuantity(ctx, ct.ID, req.ProductID, req.Quantity); err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
			return errors.MakeNotFoundError(ctx, "cart item not found")
		}
		if errors.Is(err, errors.ErrOptimisticLockConflict) {
			return errors.MakeConflictError(ctx, "cart item update conflict")
		}
		return errors.Wrap(ctx, err)
	}

	updated, err := h.cartRepo.FindByUserID(ctx, requesterID)
	if err != nil {
		return errors.Wrap(ctx, err)
	}
	return c.JSON(http.StatusOK, response.FromCart(updated))
}

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

func quantityOf(c *cart.Cart, productID string) int {
	for _, item := range c.Items {
		if item.ProductID == productID {
			return item.Quantity
		}
	}
	return 0
}
