# トランザクションルール

対象: `backend/`（Go / GORM / MySQL）。

## 基本方針

- 複数テーブルへの書き込みを含む処理は必ず `ITransactionManager.RunInTx` でラップする
- 読み取り専用操作（SELECT）はトランザクション外で実行してよい
- バリデーションと入力チェックはトランザクション開始前に全件完了させる
- **外部API呼び出し（決済・通知・KMS等）をトランザクション内に入れない**（下記）

## トランザクションを呼ぶ場所（新規コード）🔴

`architecture.md`「基本方針：レイヤードアーキテクチャ」と対応する。

- **シンプルな CRUD（usecase を作らない場合）は、ハンドラが `ITransactionManager.RunInTx` を直接呼ぶ。** usecase 層を作らない方針と整合させる
- **複雑なオーケストレーション（外部API呼び出し・暗号化などを伴い usecase 層を挟む場合）は、service 層がトランザクションを保持する**
- どちらの場合も、**外部API呼び出し・暗号化/復号処理をトランザクション内に置かない**のは変わらず絶対（下記）

## 外部API呼び出しとトランザクションを重ねない 🔴

決済API・通知送信・KMSでの暗号化など、外部サービスへの呼び出しはレイテンシが読めない。
**この間トランザクションを開いたままにすると、DB接続を長時間占有し、スケール時に接続が枯渇する。**

```go
// 誤: 外部API呼び出しをトランザクションで囲む
return s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
    result, err := s.paymentClient.Charge(txCtx, req)  // 応答時間が読めない呼び出しを保持
    ...
})

// 正: 呼び出し → その後に短いトランザクションで書き込む
result, err := s.paymentClient.Charge(ctx, req)
if err != nil {
    return errors.Wrap(ctx, err)
}
return s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
    if err := s.orderRepo.Save(txCtx, order); err != nil {
        return errors.Wrap(txCtx, err)
    }
    return s.paymentRepo.Save(txCtx, result)
})
```

暗号化（`internal/crypto/` 等、KMS呼び出しを伴うもの）も同様にトランザクション外で行う。

## インターフェースと実装の場所

- `ITransactionManager` は `internal/database/transaction.go` で定義する（`package database`。`logger.go` と同じ階層。DBエンジン非依存の共通実装で、`internal/database/mysql/` と並ぶ）
- `TransactionManager` はステートレスで、`*gorm.DB` は `ctx` から取得する（フィールドに持たない）
- モックは `moq` で自動生成し、`internal/database/mock/` に置く
- 手書きスタブ禁止

## コンストラクタでの注入

```go
func NewCartService(
    cartRepo  domaincart.ICartRepository,
    orderRepo domainorder.IOrderRepository,
    txManager transaction.ITransactionManager,
) *CartService { ... }
```

## RunInTx の実装パターン

```go
func (s *CartService) Checkout(ctx context.Context, cartID string) error {
    // バリデーション（トランザクション外）
    cart, err := s.cartRepo.FindByID(ctx, cartID)
    if err != nil {
        return errors.Wrap(ctx, err)
    }
    if err := cart.CanCheckout(); err != nil {
        return errors.MakeBusinessError(ctx, "cart cannot be checked out")
    }

    // DB書き込み（トランザクション内）
    return s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
        if err := s.orderRepo.Create(txCtx, order.FromCart(cart)); err != nil {
            return errors.Wrap(txCtx, err)
        }
        return s.cartRepo.Clear(txCtx, cartID)
    })
}
```

クロージャ外に結果を持ち出す場合は外側の変数を使う。

```go
var orderID string
err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
    o := &order.Order{...}
    if err := s.orderRepo.Create(txCtx, o); err != nil {
        return errors.Wrap(txCtx, err)
    }
    orderID = o.ID
    return nil
})
```

### ハンドラから直接呼ぶ場合（シンプルな CRUD・usecase なし）

```go
func (h *xxxHandler) Update(c *echo.Context) error {
    ctx := c.Request().Context()

    // バリデーション（トランザクション外）
    x, err := h.repo.FindByID(ctx, id)
    if err != nil {
        return errors.Wrap(ctx, err)
    }
    if err := x.CanUpdate(); err != nil {
        return errors.MakeBusinessError(ctx, "xxx cannot be updated")
    }

    // DB書き込み（トランザクション内）。楽観ロックは GORM プラグインが処理する
    if err := h.txManager.RunInTx(ctx, func(txCtx context.Context) error {
        return h.repo.Update(txCtx, x)
    }); err != nil {
        return errors.Wrap(ctx, err)
    }
    return c.JSON(http.StatusOK, response.FromXxx(x))
}
```

## スケールとコネクションの注意

- Cloud Run はスケール0から起動するため、DB接続が急増しうる（`architecture.md`「設定・初期化のルール」）。**長時間トランザクション・大量行ロックは接続枯渇に直結するため避ける**
- 同時実行下での競合解決は楽観ロック（`Version`チェック。`architecture.md`「domain＝DBモデル・BaseModel・楽観ロック」）に委ねる。悲観ロックはよほどの理由がない限り使わない
- トランザクション内の往復回数（クエリ発行回数）を最小にする

## 禁止事項

- `gorm.DB.Transaction` / `gorm.DB.Begin` をハンドラ・サービス層で直接呼ばない。`ITransactionManager` 経由のみ
- `pkg/errors` 以外のエラーパッケージを使わない
- 手書きスタブでのモック（`moq` 生成を使う）
- バリデーションをトランザクション内に混在させない
- **外部API呼び出し・暗号化/復号処理をトランザクション内に置かない**
- **GORM の `AutoMigrate` を使わない。** スキーマの正は `db/migrations/`（`sql-migrate`。`cmd/migrate`）
- **楽観ロック（`version`チェック）をアプリケーションコードで手書きしない。** GORM プラグインに委ねる（`architecture.md`「domain＝DBモデル・BaseModel・楽観ロック」）

## テストでのモック利用

```go
txMock := &transactionmock.ITransactionManagerMock{
    RunInTxFunc: func(ctx context.Context, fn func(ctx context.Context) error) error {
        return fn(ctx)  // テストでは fn をそのまま呼ぶ（トランザクションなし）
    },
}
svc := service.NewCartService(cartRepo, orderRepo, txMock)
```

## トランザクションが不要なケース

- 単一テーブルへの書き込みのみ
- 読み取りのみ
- 外部API呼び出し単体（決済・通知送信など、DB書き込みを伴わないもの）
