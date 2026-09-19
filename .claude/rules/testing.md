# テストルール

## Go（`backend/`）

### ファイル配置

- テストファイルは実装ファイルと同一パッケージ・同一ディレクトリに置く（`foo.go` → `foo_test.go`）
- ブラックボックステストが必要な場合のみ `package xxx_test` を使う

### テストの書き方

- テーブル駆動テスト（`[]struct{ name, input, want }`）を基本形とする
- サブテストは `t.Run` で命名する。テスト名は日本語可
- HTTP handler のテストは `net/http/httptest` + `echo.New()` で実際のルーターを立てる
- ドメイン層の純粋関数（`CanXxx` のようなバリデーション・状態遷移判定）を優先してテストする（`architecture.md`）

```go
func TestCart_CanCheckout(t *testing.T) {
    tests := []struct {
        name    string
        items   []cart.Item
        wantErr bool
    }{
        {"商品が入っている場合は注文できる", []cart.Item{{ProductID: "p1", Quantity: 1}}, false},
        {"カートが空の場合は注文できない", nil, true},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            c := &cart.Cart{Items: tc.items}
            err := c.CanCheckout()
            if tc.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
        })
    }
}
```

```go
func TestOrderServer_Create(t *testing.T) {
    tests := []struct {
        name       string
        body       string
        wantStatus int
    }{
        {"正常な注文", `{"items":[{"product_id":"p1","quantity":1}]}`, http.StatusCreated},
        {"空のカートで注文", `{"items":[]}`, http.StatusUnprocessableEntity},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            e := echo.New()
            req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(tc.body))
            rec := httptest.NewRecorder()
            c := e.NewContext(req, rec)
            // ...
            assert.Equal(t, tc.wantStatus, rec.Code)
        })
    }
}
```

### モック方針

- **DB は実 DB を使う**（`docker compose up -d db` で起動した MySQL）。GORM のモック禁止
- **外部サービスクライアント（`internal/external/` 等）は必ずインターフェースモックを使う。実 API を呼ばない**（課金・非決定性・CI の外部依存を避けるため）
- モックは `internal/domain/<ドメイン>/mock/` に置き、`moq` で生成する（`package-structure.md`）
- **テスト内にモック・スタブをローカル実装しない**

```go
// 正: moq 生成のモックを使う
repoMock := &moq.IProductRepositoryMock{
    FindByIDFunc: func(ctx context.Context, id string) (*product.Product, error) {
        return testProduct, nil
    },
}
```

### 禁止事項

- `time.Sleep` をテスト内で使わない
- テストで外部ネットワークにアクセスしない
- `os.Exit` を呼ぶコードはテストできないため、`main` 以外では使わない

### カバレッジ

- `go test ./... -coverprofile=coverage.out` で計測する（CI では `.github/workflows/lint_and_test.yml` が実行する）
- `domain/` のビジネスロジックは 80% 以上を目標とする

## フロントエンド（`frontend/`。Next.js）

### テスト

- テストフレームワーク: **Vitest**（`pnpm test`。実行環境は `happy-dom`）
- 対象は `vitest.config.ts` の `include`（`**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}`）に従う。ロジック（`api/`, `lib/`, `hooks/` 等の純粋関数・フック）を優先してテストする
- コンポーネントの見た目のスナップショット比較（VRT）の仕組みは未導入。追加する場合は `@testing-library/react` 等の導入から検討し、`package.json` に明記する
- 型チェックは `pnpm type-check`（tsgo）、Lint は `pnpm lint`（oxlint）で担保する。テストコード自体もこれらの対象に含める

```ts
describe('useCart', () => {
  it('商品追加時に合計金額が更新される', () => { ... })
})
```

- コミット前に `pnpm lint` / `pnpm type-check` / `pnpm test` / `pnpm format` が通ること
