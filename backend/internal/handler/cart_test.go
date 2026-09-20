package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/o-ga09/go-backend-template/internal/domain/cart"
	cartmoq "github.com/o-ga09/go-backend-template/internal/domain/cart/mock"
	"github.com/o-ga09/go-backend-template/internal/domain/product"
	productmoq "github.com/o-ga09/go-backend-template/internal/domain/product/mock"
	"github.com/o-ga09/go-backend-template/internal/handler"
	"github.com/o-ga09/go-backend-template/internal/handler/response"
	"github.com/o-ga09/go-backend-template/internal/server"
	"github.com/o-ga09/go-backend-template/pkg/errors"
)

func TestCartHandler_Get(t *testing.T) {
	tests := []struct {
		name        string
		requesterID string
		cartRepo    *cartmoq.ICartRepositoryMock
		wantStatus  int
	}{
		{
			name:        "自分のカートが取得できる",
			requesterID: "user-1",
			cartRepo: &cartmoq.ICartRepositoryMock{
				FindByUserIDFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
					c := &cart.Cart{UserID: userID, Items: []cart.CartItem{{ProductID: "p1", Quantity: 2}}}
					c.ID = "cart-1"
					return c, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "カートが無い場合は空カートを返す",
			requesterID: "user-1",
			cartRepo: &cartmoq.ICartRepositoryMock{
				FindByUserIDFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
					return nil, errors.ErrRecordNotFound
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "未認証の場合は401",
			requesterID: "",
			cartRepo:    &cartmoq.ICartRepositoryMock{},
			wantStatus:  http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, rec := newTestContext(t, http.MethodGet, "/api/cart", "", tc.requesterID)
			h := handler.NewCartHandler(tc.cartRepo, &productmoq.IProductRepositoryMock{})

			err := h.Get(c)

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

func TestCartHandler_Get_リクエスト主体の識別子で自分のカートのみ取得する(t *testing.T) {
	// カートIDをリクエストから受け取らない設計のため、requesterID以外のカートを
	// 直接指定して取得することはできない(global-constraints.md/task-3-brief.md)。
	var gotUserID string
	cartRepo := &cartmoq.ICartRepositoryMock{
		FindByUserIDFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
			gotUserID = userID
			c := &cart.Cart{UserID: userID}
			c.ID = "cart-1"
			return c, nil
		},
	}

	c, _ := newTestContext(t, http.MethodGet, "/api/cart", "", "user-1")
	h := handler.NewCartHandler(cartRepo, &productmoq.IProductRepositoryMock{})

	if err := h.Get(c); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if gotUserID != "user-1" {
		t.Errorf("FindByUserID was called with userID = %q, want %q", gotUserID, "user-1")
	}
}

// newCartNotFoundThenCreatedMockはカート未作成のユーザーがAddItemした際の
// フロー(FindByUserID:not found→Create→AddItem→FindByUserID:作成済み)を
// 再現するステートフルなモックを組み立てる。
func newCartNotFoundThenCreatedMock() *cartmoq.ICartRepositoryMock {
	findByUserIDCalls := 0
	return &cartmoq.ICartRepositoryMock{
		FindByUserIDFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
			findByUserIDCalls++
			if findByUserIDCalls == 1 {
				return nil, errors.ErrRecordNotFound
			}
			c := &cart.Cart{UserID: userID, Items: []cart.CartItem{{ProductID: "p1", Quantity: 1}}}
			c.ID = "cart-new"
			return c, nil
		},
		CreateFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
			c := &cart.Cart{UserID: userID}
			c.ID = "cart-new"
			return c, nil
		},
		AddItemFunc: func(ctx context.Context, cartID, productID string, quantity int) error {
			return nil
		},
	}
}

func TestCartHandler_AddItem(t *testing.T) {
	tests := []struct {
		name        string
		requesterID string
		body        string
		cartRepo    *cartmoq.ICartRepositoryMock
		productRepo *productmoq.IProductRepositoryMock
		wantStatus  int
	}{
		{
			name:        "既存カートに商品を追加できる",
			requesterID: "user-1",
			body:        `{"productId":"p1","quantity":2}`,
			productRepo: &productmoq.IProductRepositoryMock{
				FindByIDFunc: func(ctx context.Context, id string) (*product.Product, error) {
					p := &product.Product{Stock: 10}
					p.ID = id
					return p, nil
				},
			},
			cartRepo: &cartmoq.ICartRepositoryMock{
				FindByUserIDFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
					c := &cart.Cart{UserID: userID}
					c.ID = "cart-1"
					return c, nil
				},
				AddItemFunc: func(ctx context.Context, cartID, productID string, quantity int) error {
					return nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "カートが無い場合は新規作成してから追加する",
			requesterID: "user-1",
			body:        `{"productId":"p1","quantity":1}`,
			productRepo: &productmoq.IProductRepositoryMock{
				FindByIDFunc: func(ctx context.Context, id string) (*product.Product, error) {
					p := &product.Product{Stock: 10}
					p.ID = id
					return p, nil
				},
			},
			cartRepo:   newCartNotFoundThenCreatedMock(),
			wantStatus: http.StatusOK,
		},
		{
			name:        "存在しない商品は404",
			requesterID: "user-1",
			body:        `{"productId":"p-404","quantity":1}`,
			productRepo: &productmoq.IProductRepositoryMock{
				FindByIDFunc: func(ctx context.Context, id string) (*product.Product, error) {
					return nil, errors.ErrRecordNotFound
				},
			},
			cartRepo:   &cartmoq.ICartRepositoryMock{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:        "在庫不足の場合は422",
			requesterID: "user-1",
			body:        `{"productId":"p1","quantity":100}`,
			productRepo: &productmoq.IProductRepositoryMock{
				FindByIDFunc: func(ctx context.Context, id string) (*product.Product, error) {
					p := &product.Product{Stock: 1}
					p.ID = id
					return p, nil
				},
			},
			cartRepo:   &cartmoq.ICartRepositoryMock{},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:        "quantityが0以下の場合は422",
			requesterID: "user-1",
			body:        `{"productId":"p1","quantity":0}`,
			productRepo: &productmoq.IProductRepositoryMock{},
			cartRepo:    &cartmoq.ICartRepositoryMock{},
			wantStatus:  http.StatusUnprocessableEntity,
		},
		{
			name:        "未認証の場合は401",
			requesterID: "",
			body:        `{"productId":"p1","quantity":1}`,
			productRepo: &productmoq.IProductRepositoryMock{},
			cartRepo:    &cartmoq.ICartRepositoryMock{},
			wantStatus:  http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, rec := newTestContext(t, http.MethodPost, "/api/cart", tc.body, tc.requesterID)
			h := handler.NewCartHandler(tc.cartRepo, tc.productRepo)

			err := h.AddItem(c)

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

func TestCartHandler_UpdateItem(t *testing.T) {
	tests := []struct {
		name        string
		requesterID string
		body        string
		cartRepo    *cartmoq.ICartRepositoryMock
		wantStatus  int
	}{
		{
			name:        "数量を変更できる",
			requesterID: "user-1",
			body:        `{"productId":"p1","quantity":5}`,
			cartRepo: &cartmoq.ICartRepositoryMock{
				FindByUserIDFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
					c := &cart.Cart{UserID: userID}
					c.ID = "cart-1"
					return c, nil
				},
				UpdateItemQuantityFunc: func(ctx context.Context, cartID, productID string, quantity int) error {
					return nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "カートが無い場合は404",
			requesterID: "user-1",
			body:        `{"productId":"p1","quantity":5}`,
			cartRepo: &cartmoq.ICartRepositoryMock{
				FindByUserIDFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
					return nil, errors.ErrRecordNotFound
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:        "対象の明細が無い場合は404",
			requesterID: "user-1",
			body:        `{"productId":"p1","quantity":5}`,
			cartRepo: &cartmoq.ICartRepositoryMock{
				FindByUserIDFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
					c := &cart.Cart{UserID: userID}
					c.ID = "cart-1"
					return c, nil
				},
				UpdateItemQuantityFunc: func(ctx context.Context, cartID, productID string, quantity int) error {
					return errors.ErrRecordNotFound
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:        "未認証の場合は401",
			requesterID: "",
			body:        `{"productId":"p1","quantity":5}`,
			cartRepo:    &cartmoq.ICartRepositoryMock{},
			wantStatus:  http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, rec := newTestContext(t, http.MethodPut, "/api/cart", tc.body, tc.requesterID)
			h := handler.NewCartHandler(tc.cartRepo, &productmoq.IProductRepositoryMock{})

			err := h.UpdateItem(c)

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

func TestCartHandler_RemoveItem(t *testing.T) {
	tests := []struct {
		name        string
		requesterID string
		query       string
		cartRepo    *cartmoq.ICartRepositoryMock
		wantStatus  int
	}{
		{
			name:        "商品を削除できる",
			requesterID: "user-1",
			query:       "productId=p1",
			cartRepo: &cartmoq.ICartRepositoryMock{
				FindByUserIDFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
					c := &cart.Cart{UserID: userID}
					c.ID = "cart-1"
					return c, nil
				},
				RemoveItemFunc: func(ctx context.Context, cartID, productID string) error {
					return nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "カートが無い場合は404",
			requesterID: "user-1",
			query:       "productId=p1",
			cartRepo: &cartmoq.ICartRepositoryMock{
				FindByUserIDFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
					return nil, errors.ErrRecordNotFound
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:        "未認証の場合は401",
			requesterID: "",
			query:       "productId=p1",
			cartRepo:    &cartmoq.ICartRepositoryMock{},
			wantStatus:  http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, rec := newTestContext(t, http.MethodDelete, "/api/cart?"+tc.query, "", tc.requesterID)
			h := handler.NewCartHandler(tc.cartRepo, &productmoq.IProductRepositoryMock{})

			err := h.RemoveItem(c)

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

// レスポンスの中身も一件だけ確認しておく(response.FromCartの変換確認)。
func TestCartHandler_Get_ResponseBody(t *testing.T) {
	cartRepo := &cartmoq.ICartRepositoryMock{
		FindByUserIDFunc: func(ctx context.Context, userID string) (*cart.Cart, error) {
			c := &cart.Cart{UserID: userID, Items: []cart.CartItem{{ProductID: "p1", Quantity: 2}}}
			c.ID = "cart-1"
			return c, nil
		},
	}
	c, rec := newTestContext(t, http.MethodGet, "/api/cart", "", "user-1")
	h := handler.NewCartHandler(cartRepo, &productmoq.IProductRepositoryMock{})

	if err := h.Get(c); err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	var got response.Cart
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response body: %v (body=%s)", err, rec.Body.String())
	}
	want := response.Cart{ID: "cart-1", Items: []response.CartItem{{ProductID: "p1", Quantity: 2}}}
	if got.ID != want.ID || len(got.Items) != len(want.Items) || got.Items[0] != want.Items[0] {
		t.Errorf("response body = %+v, want %+v", got, want)
	}
}
