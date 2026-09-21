package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/labstack/echo/v5"

	"github.com/o-ga09/go-backend-template/internal/domain/product"
	moq "github.com/o-ga09/go-backend-template/internal/domain/product/mock"
	"github.com/o-ga09/go-backend-template/internal/handler"
	"github.com/o-ga09/go-backend-template/internal/handler/response"
	"github.com/o-ga09/go-backend-template/internal/server"
	"github.com/o-ga09/go-backend-template/pkg/errors"
)

func TestProductHandler_List(t *testing.T) {
	tests := []struct {
		name       string
		repo       *moq.IProductRepositoryMock
		wantStatus int
		wantBody   []response.Product
	}{
		{
			name: "商品一覧が取得できる",
			repo: &moq.IProductRepositoryMock{
				ListFunc: func(ctx context.Context) ([]*product.Product, error) {
					p := &product.Product{Name: "mug", Description: "彩り マグカップ", PriceYen: 1000, Stock: 5}
					p.ID = "product-1"
					return []*product.Product{p}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody: []response.Product{
				{ID: "product-1", Name: "mug", Description: "彩り マグカップ", PriceYen: 1000, Stock: 5},
			},
		},
		{
			name: "商品0件の場合はnullではなく空配列が返る",
			repo: &moq.IProductRepositoryMock{
				ListFunc: func(ctx context.Context) ([]*product.Product, error) {
					return []*product.Product{}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   []response.Product{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, rec := newTestContext(t, http.MethodGet, "/api/products", "", "")
			h := handler.NewProductHandler(tc.repo)

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

			if body := rec.Body.String(); body == "null\n" || body == "null" {
				t.Fatalf("response body must not be null (want []), got %q", body)
			}

			var got []response.Product
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("failed to decode response body: %v (body=%s)", err, rec.Body.String())
			}
			if len(got) != len(tc.wantBody) {
				t.Fatalf("response body length = %d, want %d (got=%+v)", len(got), len(tc.wantBody), got)
			}
			for i, want := range tc.wantBody {
				if got[i] != want {
					t.Errorf("response body[%d] = %+v, want %+v", i, got[i], want)
				}
			}
		})
	}
}

func TestProductHandler_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		targetID   string
		repo       *moq.IProductRepositoryMock
		wantStatus int
		wantBody   *response.Product
	}{
		{
			name:     "存在する商品は取得できる",
			targetID: "product-1",
			repo: &moq.IProductRepositoryMock{
				FindByIDFunc: func(ctx context.Context, id string) (*product.Product, error) {
					p := &product.Product{Name: "mug", Description: "彩り マグカップ", PriceYen: 1000, Stock: 5}
					p.ID = id
					return p, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   &response.Product{ID: "product-1", Name: "mug", Description: "彩り マグカップ", PriceYen: 1000, Stock: 5},
		},
		{
			name:     "存在しない商品は404",
			targetID: "product-404",
			repo: &moq.IProductRepositoryMock{
				FindByIDFunc: func(ctx context.Context, id string) (*product.Product, error) {
					return nil, errors.ErrRecordNotFound
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, rec := newTestContext(t, http.MethodGet, "/api/products/:id", "", "")
			c.SetPathValues(echo.PathValues{{Name: "id", Value: tc.targetID}})

			h := handler.NewProductHandler(tc.repo)
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

			if tc.wantBody != nil {
				var got response.Product
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("failed to decode response body: %v (body=%s)", err, rec.Body.String())
				}
				if got != *tc.wantBody {
					t.Errorf("response body = %+v, want %+v", got, *tc.wantBody)
				}
			}
		})
	}
}
