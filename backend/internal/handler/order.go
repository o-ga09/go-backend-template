package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/o-ga09/go-backend-template/internal/database"
	"github.com/o-ga09/go-backend-template/internal/domain/cart"
	"github.com/o-ga09/go-backend-template/internal/domain/order"
	"github.com/o-ga09/go-backend-template/internal/domain/product"
	"github.com/o-ga09/go-backend-template/internal/handler/request"
	"github.com/o-ga09/go-backend-template/internal/handler/response"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	"github.com/o-ga09/go-backend-template/pkg/errors"
)

// 注文操作は全て認証必須で、常にログイン中ユーザー本人の注文のみを対象とする。
type orderHandler struct {
	orderRepo   order.IOrderRepository
	cartRepo    cart.ICartRepository
	productRepo product.IProductRepository
	txManager   database.ITransactionManager
}

type IOrder interface {
	Create(c *echo.Context) error
	List(c *echo.Context) error
	GetByID(c *echo.Context) error
}

func NewOrderHandler(orderRepo order.IOrderRepository, cartRepo cart.ICartRepository, productRepo product.IProductRepository, txManager database.ITransactionManager) IOrder {
	return &orderHandler{orderRepo: orderRepo, cartRepo: cartRepo, productRepo: productRepo, txManager: txManager}
}

// 商品検索・バリデーションはトランザクション外で行い、書き込みとカートのクリアのみを
// RunInTxでラップする。
// POST /api/orders
func (h *orderHandler) Create(c *echo.Context) error {
	ctx := c.Request().Context()

	requesterID := Ctx.GetUserID(ctx)
	if requesterID == "" {
		return errors.MakeAuthorizedError(ctx, "authentication required")
	}

	ct, err := h.cartRepo.FindByUserID(ctx, requesterID)
	if err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
			return errors.MakeBusinessError(ctx, "cart is empty")
		}
		return errors.Wrap(ctx, err)
	}
	if err := ct.CanCheckout(); err != nil {
		if errors.Is(err, errors.ErrCartEmpty) {
			return errors.MakeBusinessError(ctx, "cart is empty")
		}
		return errors.Wrap(ctx, err)
	}

	products := make(map[string]*product.Product, len(ct.Items))
	for _, item := range ct.Items {
		p, err := h.productRepo.FindByID(ctx, item.ProductID)
		if err != nil {
			if errors.Is(err, errors.ErrRecordNotFound) {
				return errors.MakeNotFoundError(ctx, "product not found")
			}
			return errors.Wrap(ctx, err)
		}
		products[item.ProductID] = p
	}

	o, err := order.NewFromCart(ct, products)
	if err != nil {
		if errors.Is(err, errors.ErrInsufficientStock) {
			return errors.MakeBusinessError(ctx, "insufficient stock")
		}
		if errors.Is(err, errors.ErrRecordNotFound) {
			return errors.MakeNotFoundError(ctx, "product not found")
		}
		return errors.Wrap(ctx, err)
	}

	if err := h.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		if err := h.orderRepo.Create(txCtx, o); err != nil {
			return err
		}
		for _, item := range o.Items {
			if err := h.productRepo.DecreaseStock(txCtx, item.ProductID, item.Quantity); err != nil {
				return err
			}
		}
		return h.cartRepo.Clear(txCtx, ct.ID)
	}); err != nil {
		if errors.Is(err, errors.ErrInsufficientStock) {
			// 事前チェック(NewFromCart)後、書き込み直前に他の注文が在庫を
			// 消費した競合。422(バリデーション違反)ではなく409(競合)として扱う。
			return errors.MakeConflictError(ctx, "insufficient stock")
		}
		return errors.Wrap(ctx, err)
	}

	return c.JSON(http.StatusCreated, response.FromOrder(o))
}

// GET /api/orders
func (h *orderHandler) List(c *echo.Context) error {
	ctx := c.Request().Context()

	requesterID := Ctx.GetUserID(ctx)
	if requesterID == "" {
		return errors.MakeAuthorizedError(ctx, "authentication required")
	}

	os, err := h.orderRepo.ListByUserID(ctx, requesterID)
	if err != nil {
		return errors.Wrap(ctx, err)
	}

	return c.JSON(http.StatusOK, response.FromOrders(os))
}

// GET /api/orders/:id
func (h *orderHandler) GetByID(c *echo.Context) error {
	ctx := c.Request().Context()

	requesterID := Ctx.GetUserID(ctx)
	if requesterID == "" {
		return errors.MakeAuthorizedError(ctx, "authentication required")
	}

	var req request.GetOrderRequest
	if err := c.Bind(&req); err != nil {
		return errors.Wrap(ctx, err)
	}
	if err := c.Validate(&req); err != nil {
		return errors.Wrap(ctx, err)
	}

	o, err := h.orderRepo.FindByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
			return errors.MakeNotFoundError(ctx, "order not found")
		}
		return errors.Wrap(ctx, err)
	}
	if !o.IsOwnedBy(requesterID) {
		return errors.MakeAuthorizationError(ctx, "cannot access other user's order")
	}

	return c.JSON(http.StatusOK, response.FromOrder(o))
}
