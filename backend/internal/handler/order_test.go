package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/labstack/echo/v5"

	"github.com/o-ga09/go-backend-template/internal/database"
	txmock "github.com/o-ga09/go-backend-template/internal/database/mock"
	"github.com/o-ga09/go-backend-template/internal/domain/cart"
	cartmoq "github.com/o-ga09/go-backend-template/internal/domain/cart/mock"
	"github.com/o-ga09/go-backend-template/internal/domain/order"
	ordermoq "github.com/o-ga09/go-backend-template/internal/domain/order/mock"
	"github.com/o-ga09/go-backend-template/internal/domain/product"
	productmoq "github.com/o-ga09/go-backend-template/internal/domain/product/mock"
	"github.com/o-ga09/go-backend-template/internal/handler"
	"github.com/o-ga09/go-backend-template/internal/handler/response"
	"github.com/o-ga09/go-backend-template/internal/server"
	"github.com/o-ga09/go-backend-template/pkg/errors"
)

func newOrderNoopTxManager() database.ITransactionManager {
	return &txmock.ITransactionManagerMock{
		RunInTxFunc: func(ctx context.Context, fn func(ctx context.Context) error) error {
			return fn(ctx)
		},
	}
}

func TestOrderHandler_Create(t *testing.T) {
	tests := []struct {
		name        string
		requesterID string
		cartRepo    *cartmoq.ICartRepositoryMock
		productRepo *productmoq.IProductRepositoryMock
		orderRepo   *ordermoq.IOrderRepositoryMock
		wantStatus  int
		wantBody    *response.Order
	}{
		{
			name:        "カートから注文を確定できる",
			requesterID: "user-1",
			cartRepo: &cartmoq.ICartRepositoryMock{
				FindByUserIDFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
					c := &cart.Cart{UserID: userID, Items: []cart.CartItem{{ProductID: "p1", Quantity: 2}}}
					c.ID = "cart-1"
					return c, nil
				},
				ClearFunc: func(ctx context.Context, cartID string) error { return nil },
			},
			productRepo: &productmoq.IProductRepositoryMock{
				FindByIDFunc: func(ctx context.Context, id string) (*product.Product, error) {
					p := &product.Product{PriceYen: 1000, Stock: 10}
					p.ID = id
					return p, nil
				},
				DecreaseStockFunc: func(ctx context.Context, productID string, quantity int) error { return nil },
			},
			orderRepo: &ordermoq.IOrderRepositoryMock{
				CreateFunc: func(ctx context.Context, o *order.Order) error {
					o.ID = "order-1"
					return nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody: &response.Order{
				ID:            "order-1",
				Status:        "pending",
				TotalPriceYen: 2000,
				Items: []response.OrderItem{
					{ProductID: "p1", Quantity: 2, UnitPriceYen: 1000},
				},
			},
		},
		{
			name:        "カートが未作成の場合は422",
			requesterID: "user-1",
			cartRepo: &cartmoq.ICartRepositoryMock{
				FindByUserIDFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
					return nil, errors.ErrRecordNotFound
				},
			},
			productRepo: &productmoq.IProductRepositoryMock{},
			orderRepo:   &ordermoq.IOrderRepositoryMock{},
			wantStatus:  http.StatusUnprocessableEntity,
		},
		{
			name:        "カートが空の場合は422",
			requesterID: "user-1",
			cartRepo: &cartmoq.ICartRepositoryMock{
				FindByUserIDFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
					c := &cart.Cart{UserID: userID}
					c.ID = "cart-1"
					return c, nil
				},
			},
			productRepo: &productmoq.IProductRepositoryMock{},
			orderRepo:   &ordermoq.IOrderRepositoryMock{},
			wantStatus:  http.StatusUnprocessableEntity,
		},
		{
			name:        "カート内商品が在庫不足の場合は422",
			requesterID: "user-1",
			cartRepo: &cartmoq.ICartRepositoryMock{
				FindByUserIDFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
					c := &cart.Cart{UserID: userID, Items: []cart.CartItem{{ProductID: "p1", Quantity: 100}}}
					c.ID = "cart-1"
					return c, nil
				},
			},
			productRepo: &productmoq.IProductRepositoryMock{
				FindByIDFunc: func(ctx context.Context, id string) (*product.Product, error) {
					p := &product.Product{PriceYen: 1000, Stock: 1}
					p.ID = id
					return p, nil
				},
			},
			orderRepo:  &ordermoq.IOrderRepositoryMock{},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:        "カート内商品が削除済みの場合は404",
			requesterID: "user-1",
			cartRepo: &cartmoq.ICartRepositoryMock{
				FindByUserIDFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
					c := &cart.Cart{UserID: userID, Items: []cart.CartItem{{ProductID: "p-404", Quantity: 1}}}
					c.ID = "cart-1"
					return c, nil
				},
			},
			productRepo: &productmoq.IProductRepositoryMock{
				FindByIDFunc: func(ctx context.Context, id string) (*product.Product, error) {
					return nil, errors.ErrRecordNotFound
				},
			},
			orderRepo:  &ordermoq.IOrderRepositoryMock{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:        "事前チェック後に在庫が競合で不足した場合は409",
			requesterID: "user-1",
			cartRepo: &cartmoq.ICartRepositoryMock{
				FindByUserIDFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
					c := &cart.Cart{UserID: userID, Items: []cart.CartItem{{ProductID: "p1", Quantity: 1}}}
					c.ID = "cart-1"
					return c, nil
				},
			},
			productRepo: &productmoq.IProductRepositoryMock{
				FindByIDFunc: func(ctx context.Context, id string) (*product.Product, error) {
					p := &product.Product{PriceYen: 1000, Stock: 1}
					p.ID = id
					return p, nil
				},
				DecreaseStockFunc: func(ctx context.Context, productID string, quantity int) error {
					return errors.ErrInsufficientStock
				},
			},
			orderRepo: &ordermoq.IOrderRepositoryMock{
				CreateFunc: func(ctx context.Context, o *order.Order) error {
					o.ID = "order-1"
					return nil
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:        "未認証の場合は401",
			requesterID: "",
			cartRepo:    &cartmoq.ICartRepositoryMock{},
			productRepo: &productmoq.IProductRepositoryMock{},
			orderRepo:   &ordermoq.IOrderRepositoryMock{},
			wantStatus:  http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, rec := newTestContext(t, http.MethodPost, "/api/orders", "", tc.requesterID)
			h := handler.NewOrderHandler(tc.orderRepo, tc.cartRepo, tc.productRepo, newOrderNoopTxManager())

			err := h.Create(c)

			if err != nil {
				status, _ := server.ErrCodeToStatusAndMessage(err)
				if status != tc.wantStatus {
					t.Fatalf("error status = %d, want %d (err=%v)", status, tc.wantStatus, err)
				}
				return
			}
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}

			if tc.wantBody != nil {
				var got response.Order
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("failed to decode response body: %v (body=%s)", err, rec.Body.String())
				}
				if !reflect.DeepEqual(got, *tc.wantBody) {
					t.Errorf("response body = %+v, want %+v", got, *tc.wantBody)
				}
			}
		})
	}
}

// 複数テーブルへの書き込み(orders+order_items+カートクリア)が
// ITransactionManager.RunInTx経由で行われ、カートがCreate後にClearされることを確認する。
func TestOrderHandler_Create_トランザクション内でorder作成とカートクリアが行われる(t *testing.T) {
	cartRepo := &cartmoq.ICartRepositoryMock{
		FindByUserIDFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
			c := &cart.Cart{UserID: userID, Items: []cart.CartItem{{ProductID: "p1", Quantity: 1}}}
			c.ID = "cart-1"
			return c, nil
		},
		ClearFunc: func(ctx context.Context, cartID string) error { return nil },
	}
	productRepo := &productmoq.IProductRepositoryMock{
		FindByIDFunc: func(ctx context.Context, id string) (*product.Product, error) {
			p := &product.Product{PriceYen: 1000, Stock: 10}
			p.ID = id
			return p, nil
		},
		DecreaseStockFunc: func(ctx context.Context, productID string, quantity int) error { return nil },
	}
	orderRepo := &ordermoq.IOrderRepositoryMock{
		CreateFunc: func(ctx context.Context, o *order.Order) error {
			o.ID = "order-1"
			return nil
		},
	}

	txCalled := false
	txManager := &txmock.ITransactionManagerMock{
		RunInTxFunc: func(ctx context.Context, fn func(ctx context.Context) error) error {
			txCalled = true
			return fn(ctx)
		},
	}

	c, _ := newTestContext(t, http.MethodPost, "/api/orders", "", "user-1")
	h := handler.NewOrderHandler(orderRepo, cartRepo, productRepo, txManager)

	if err := h.Create(c); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !txCalled {
		t.Fatal("RunInTx was not called")
	}

	createCalls := orderRepo.CreateCalls()
	if len(createCalls) != 1 {
		t.Fatalf("Create calls = %d, want 1", len(createCalls))
	}
	clearCalls := cartRepo.ClearCalls()
	if len(clearCalls) != 1 || clearCalls[0].CartID != "cart-1" {
		t.Fatalf("Clear calls = %+v, want single call with cartID=cart-1", clearCalls)
	}
	decreaseStockCalls := productRepo.DecreaseStockCalls()
	if len(decreaseStockCalls) != 1 || decreaseStockCalls[0].ProductID != "p1" || decreaseStockCalls[0].Quantity != 1 {
		t.Fatalf("DecreaseStock calls = %+v, want single call with productID=p1, quantity=1", decreaseStockCalls)
	}
}

func TestOrderHandler_List(t *testing.T) {
	tests := []struct {
		name        string
		requesterID string
		orderRepo   *ordermoq.IOrderRepositoryMock
		wantStatus  int
		wantBody    []response.Order
	}{
		{
			name:        "自分の注文履歴が取得できる",
			requesterID: "user-1",
			orderRepo: &ordermoq.IOrderRepositoryMock{
				ListByUserIDFunc: func(ctx context.Context, userID string) ([]*order.Order, error) {
					o := &order.Order{
						UserID:        userID,
						Status:        order.StatusPending,
						TotalPriceYen: 500,
						Items:         []order.OrderItem{{ProductID: "p1", Quantity: 1, UnitPriceYen: 500}},
					}
					o.ID = "order-1"
					return []*order.Order{o}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody: []response.Order{
				{
					ID:            "order-1",
					Status:        "pending",
					TotalPriceYen: 500,
					Items:         []response.OrderItem{{ProductID: "p1", Quantity: 1, UnitPriceYen: 500}},
				},
			},
		},
		{
			name:        "未認証の場合は401",
			requesterID: "",
			orderRepo:   &ordermoq.IOrderRepositoryMock{},
			wantStatus:  http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, rec := newTestContext(t, http.MethodGet, "/api/orders", "", tc.requesterID)
			h := handler.NewOrderHandler(tc.orderRepo, &cartmoq.ICartRepositoryMock{}, &productmoq.IProductRepositoryMock{}, newOrderNoopTxManager())

			err := h.List(c)

			if err != nil {
				status, _ := server.ErrCodeToStatusAndMessage(err)
				if status != tc.wantStatus {
					t.Fatalf("error status = %d, want %d (err=%v)", status, tc.wantStatus, err)
				}
				return
			}
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}

			if tc.wantBody != nil {
				var got []response.Order
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("failed to decode response body: %v (body=%s)", err, rec.Body.String())
				}
				if !reflect.DeepEqual(got, tc.wantBody) {
					t.Errorf("response body = %+v, want %+v", got, tc.wantBody)
				}
			}
		})
	}
}

// 自分のユーザーID以外を指定して他ユーザーの注文履歴を取得することはできない
// (requesterID以外をクエリ等で受け取らない設計)。
func TestOrderHandler_List_リクエスト主体の識別子で自分の注文のみ取得する(t *testing.T) {
	orderRepo := &ordermoq.IOrderRepositoryMock{
		ListByUserIDFunc: func(ctx context.Context, userID string) ([]*order.Order, error) {
			return []*order.Order{}, nil
		},
	}

	c, _ := newTestContext(t, http.MethodGet, "/api/orders", "", "user-1")
	h := handler.NewOrderHandler(orderRepo, &cartmoq.ICartRepositoryMock{}, &productmoq.IProductRepositoryMock{}, newOrderNoopTxManager())

	if err := h.List(c); err != nil {
		t.Fatalf("List() error = %v", err)
	}

	calls := orderRepo.ListByUserIDCalls()
	if len(calls) != 1 || calls[0].UserID != "user-1" {
		t.Fatalf("ListByUserID calls = %+v, want single call with userID=user-1", calls)
	}
}

func TestOrderHandler_GetByID(t *testing.T) {
	tests := []struct {
		name        string
		requesterID string
		targetID    string
		orderRepo   *ordermoq.IOrderRepositoryMock
		wantStatus  int
	}{
		{
			name:        "自分の注文は取得できる",
			requesterID: "user-1",
			targetID:    "order-1",
			orderRepo: &ordermoq.IOrderRepositoryMock{
				FindByIDFunc: func(ctx context.Context, id string) (*order.Order, error) {
					o := &order.Order{UserID: "user-1"}
					o.ID = id
					return o, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "他人の注文は403",
			requesterID: "user-1",
			targetID:    "order-2",
			orderRepo: &ordermoq.IOrderRepositoryMock{
				FindByIDFunc: func(ctx context.Context, id string) (*order.Order, error) {
					o := &order.Order{UserID: "user-2"}
					o.ID = id
					return o, nil
				},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:        "存在しない注文は404",
			requesterID: "user-1",
			targetID:    "order-404",
			orderRepo: &ordermoq.IOrderRepositoryMock{
				FindByIDFunc: func(ctx context.Context, id string) (*order.Order, error) {
					return nil, errors.ErrRecordNotFound
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:        "未認証の場合は401",
			requesterID: "",
			targetID:    "order-1",
			orderRepo:   &ordermoq.IOrderRepositoryMock{},
			wantStatus:  http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, rec := newTestContext(t, http.MethodGet, "/api/orders/:id", "", tc.requesterID)
			c.SetPathValues(echo.PathValues{{Name: "id", Value: tc.targetID}})
			h := handler.NewOrderHandler(tc.orderRepo, &cartmoq.ICartRepositoryMock{}, &productmoq.IProductRepositoryMock{}, newOrderNoopTxManager())

			err := h.GetByID(c)

			if err != nil {
				status, _ := server.ErrCodeToStatusAndMessage(err)
				if status != tc.wantStatus {
					t.Fatalf("error status = %d, want %d (err=%v)", status, tc.wantStatus, err)
				}
				return
			}
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}
