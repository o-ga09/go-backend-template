# アーキテクチャルール

対象: `backend/`（Go / Echo / GORM）。クライアント（Next.js）は `frontend.md`。

**本ルールは実装の作法を決めるものであり、個別のプロダクト要件・ADR があればそちらを優先する。**

## 全体構成

```
[Next.js クライアント]
        │ HTTPS
[Echo（Cloud Run または ECS）]
        │ GORM
[MySQL]
```

## 基本方針：レイヤードアーキテクチャ 🔴

基本形は2層（`server` → `domain`）。usecase 層（`internal/service/`）を挟むのは、複数リポジトリ・外部API呼び出し・暗号化など**複雑なオーケストレーションが必要な場合のみ**の例外。

```
[シンプルな CRUD（デフォルト）]
server (handler) → domain（repository interface + ドメインロジック）

[複雑なオーケストレーション（外部API呼び出し・暗号化・複数ドメイン合成など。例外）]
server (handler) → service (usecase) → domain
```

- **シンプルな CRUD に usecase を作らない。** ハンドラが直接 domain のリポジトリを呼ぶ
- **ドメインロジック（バリデーション・状態遷移の可否判定）は必ず domain 層の純粋関数に置く。** ハンドラ・usecase はそれを呼び出すだけにする
- **domain のモデルは DB のモデルを兼ねる。** GORM 用の変換専用構造体を作らない。暗号化やカラム名の詰め替えが必要なドメインだけ、変換用の構造体を用意してよい（詳細は下記「domain＝DBモデル・BaseModel」）
- リポジトリ interface は domain 層に置く（`entity.go` の `go:generate moq`）。usecase を挟む場合も、リポジトリ interface の置き場所はドメイン層のまま変わらない

## レイヤー構成と依存方向

依存は必ず内側に向ける。外側のレイヤーが内側を知る。内側は外側を知らない。

```
[シンプルな CRUD]
server            → domain, database, database/mysql （配線と直接呼び出し）
database/mysql    → domain, database   (ドメイン型・ITransactionManagerを使うが、ドメインロジックを持たない)

[複雑なオーケストレーション]
server            → service
service           → domain
service           → database, crypto, external/<name>
external/<name>   → domain  (入出力にドメイン型を使う。外部SDKの型を漏らさない)
crypto            → 依存なし  (暗号化/復号のみ。ドメインを知らない)
```

- `domain/` は `server/`, `service/`, `database/`, `external/` を import しない
- **`internal/service/` は複雑なオーケストレーションが必要な場合のみのユースケース層。** ハンドラ（`server/`）から呼ばれ、複数のリポジトリ・`crypto.ISealer`・`ITransactionManager`・外部APIクライアントを組み合わせる。ドメインルール自体（バリデーション・状態遷移の可否判定）はドメイン層の純粋関数に置き、service はそれを呼び出すだけにする
- `database/mysql/` はビジネスロジックを持たない（データの読み書きのみ）
- `external/<name>/` は外部SDK（決済・通知など）をラップする。**SDK 型を domain や server に漏らさない**
- **暗号化・復号はトランザクションの外で行う。** `database/mysql/` のリポジトリ内で `crypto.ISealer` を呼ばない（`transaction.md`「外部API呼び出しとトランザクションを重ねない」と同じ理由で、KMS呼び出しもトランザクション内に置かない）。ドメインエンティティが保持するのは暗号文（`pkg/secret.Sealed`）で、平文（`pkg/secret.Text`）は呼び出し元（ハンドラ、または service 層）のローカル変数としてのみ生存する
- あるドメインが別ドメインの状態を参照する必要がある場面（例: 注文ドメインが在庫ドメインの在庫数を必要とする）は、**ドメイン間に新しい矢印を作らない。** 呼び出し側（ハンドラ、または service 層）が相手ドメインの参照用インターフェースを呼び、判定結果だけを対象ドメインの純粋関数（例: `order.CanPlace(stock int)`）に渡す

## ドメイン層のルール

- エンティティは `internal/domain/<ドメイン名>/entity.go` に置く（1ドメイン1パッケージ。`package-structure.md`）
- バリデーションロジックはドメイン層のメソッドとして実装する
- ドメイン間の依存は最小限にする。循環 import 禁止

```go
// 正: ドメイン型のメソッドとしてバリデーション
func (c *Cart) CanCheckout() error { ... }

// 誤: handler 層でバリデーション
if len(c.Items) == 0 { ... }
```

現在の実装例: `user`（`internal/domain/user/`）。新規ドメイン（EC商材の商品・カート・注文など）を追加する場合も、同じ2層構成に従う。

## 認可（所有者ベース）のパターン 🔴

ログイン必須・本人のリソースのみアクセス可、という認可はハンドラ + domain + `pkg/authz` の組み合わせで実装する（`internal/handler/user.go` の `GetByID` が実装例）。

1. **未ログイン（401）**: ハンドラの先頭で `Ctx.GetUserID(ctx)`（`pkg/context`）が空文字かどうかを見る。空なら `errors.MakeAuthorizedError(ctx, "authentication required")` を返す。`Authenticate` ミドルウェア（`internal/server/middleware.go`）はセッションCookieが無効/無い場合でもリクエストを拒否せず素通しするため、ログイン必須の判定は**必ずハンドラ側**で行う（`context-propagation.md`「認可ミドルウェア／ドメイン層で明示的に検証する」）
2. **リソース取得**: `requesterID` を使わず対象リソースをリポジトリから取得する。見つからなければ通常どおり `errors.MakeNotFoundError`（404）
3. **本人確認（403）**: 取得したリソースの domain メソッド（例: `(*User).IsOwnedBy(requesterID string) bool`）で所有者チェックする。この判定自体は `pkg/authz.IsOwner(requesterID, ownerID)` に委譲し、ハンドラはそのbool結果からエラーへの変換のみ行う

```go
// internal/domain/user/user.go
func (u *User) IsOwnedBy(requesterID string) bool {
    return authz.IsOwner(requesterID, u.ID)
}

// internal/handler/user.go
requesterID := Ctx.GetUserID(ctx)
if requesterID == "" {
    return errors.MakeAuthorizedError(ctx, "authentication required")
}
u, err := h.repo.FindByID(ctx, req.ID)
if err != nil {
    if errors.Is(err, errors.ErrRecordNotFound) {
        return errors.MakeNotFoundError(ctx, "user not found")
    }
    return errors.Wrap(ctx, err)
}
if !u.IsOwnedBy(requesterID) {
    return errors.MakeAuthorizationError(ctx, "cannot access other user's resource")
}
```

新規ドメイン（cart/order等）で所有者ベースの認可が必要な場合も、`pkg/authz.IsOwner` を再利用し、ドメインごとに認可ロジックを再実装しない。

## domain＝DBモデル・BaseModel・楽観ロック 🔴

- `ID` / `Version` / `CreatedAt` / `UpdatedAt` など必須カラムは、共通の `BaseModel` を domain 構造体に埋め込んで持つ

  ```go
  // pkg/model/base.go
  type BaseModel struct {
      ID        string
      Version   int
      CreatedAt time.Time
      UpdatedAt time.Time
  }

  // internal/domain/xxx/entity.go
  type Xxx struct {
      model.BaseModel
      // ドメイン固有フィールド
  }
  ```

- 監査項目（作成者など）が必要なドメインは `BaseModel` に `CreatedUserID` 等を追加する。個別ドメインに都度同じフィールドを生やさない
- **`BaseModel` のカラム（ID採番・`CreatedAt`/`UpdatedAt`・`Version`のインクリメント）は GORM プラグイン（`Callbacks`）で自動挿入する。** リポジトリや変換関数の中で代入しない
- **楽観ロック（`Version`チェック）は GORM のプラグイン機構で実装する。** `WHERE version = ?` をリポジトリに手書きしたり、アプリケーションコードで比較・分岐を書かない。競合時は GORM が返すエラーをリポジトリ層で判別し、`errors.MakeConflictError` に変換する
- 暗号化やカラム名の詰め替えが必要なドメインは、この前提から外れてよい（従来通りリポジトリで手動マッピングする）

これらは domain がそのまま GORM モデルとして永続化される前提（domain＝DBモデル）で成り立つ。GORM プラグイン本体は `internal/database/mysql/` に置く。

## データベース層のルール

- DB アクセスは `internal/database/mysql/` 配下の GORM リポジトリ経由のみ
- 各ドメインの CRUD は個別ファイルに分離する（`auth.go`, `order.go` など）
- 外部サービスクライアントはインターフェース越しに呼ぶ（例: 決済API → `internal/external/payment/`、Cloud KMS → `internal/crypto/`）
- **マイグレーションは GORM の AutoMigrate を使わない。** `db/migrations/` の SQL と `cmd/migrate`（`sql-migrate`）で管理する。スキーマの正はマイグレーションファイル（`BaseModel` の各カラムも例外なくマイグレーションSQLで定義する）

## 設定・初期化のルール

- 環境変数は `pkg/config/config.go` の `Config` 構造体で一元管理する（`caarlos0/env`）
- `Config` は起動時に context に格納し、`context.GetCfgFromCtx(ctx)` で取得する
- DB クライアント・外部サービスクライアントはコンストラクタで直接注入する
- **Cloud Run はスケール0から起動するため、DB接続が急増しうる。** `gorm.DB` から `*sql.DB` を取り出して `SetMaxOpenConns` / `SetMaxIdleConns` / `SetConnMaxLifetime` を起動時に必ず設定する

## 依存注入のルール

コンストラクタで受け取った依存はメソッド内で nil チェックしない。

このプロジェクトが使う echo v5（`github.com/labstack/echo/v5`）の `Context` は v4 以前と異なり **interface ではなく struct** なので、ハンドラ・ミドルウェアのシグネチャは `func(c *echo.Context) error`（`c echo.Context` ではなく `c *echo.Context`）になる。

```go
// 誤: メソッド内に nil ガードを書く
func (h *orderHandler) Create(c *echo.Context) error {
    if h.repo == nil {
        return nil
    }
    ...
}

// 正: コンストラクタで依存を受け取り、メソッドはそのまま使う
func (h *orderHandler) Create(c *echo.Context) error {
    res, err := h.repo.Create(ctx, order)
    ...
}
```

- `if h.repo != nil`, `if h.client != nil` のような防御的チェックを実装メソッドに書かない
- 必須の依存はコンストラクタ引数で明示し、nil を渡せないようにする
- テスト用モックはインターフェースから `moq` で自動生成したものを使う

### ハンドラも interface 越しに公開する 🔴

- ハンドラ構造体は非公開（`orderHandler`）にし、そのハンドラが実装する公開インターフェース（`IOrder`）を `internal/handler/<name>.go` に定義する。`NewOrderHandler` はそのインターフェースを返す（domain のリポジトリ interface と同じ形。`package-structure.md`「interface の置き場所」）
- **インターフェース名に `Handler` サフィックスを付けない。** `IOrderHandler` にすると呼び出し側で `handler.IOrderHandler` のようにパッケージ名（`handler`）と `Handler` が重複する。パッケージ名で「これはハンドラである」ことが分かるため、インターフェース名は対象リソース名（`IOrder`, `IUser`, `IAuth` 等）だけにする

```go
// internal/handler/order.go
type orderHandler struct {
    repo order.IOrderRepository
}

type IOrder interface {
    Create(c *echo.Context) error
    GetByID(c *echo.Context) error
}

func NewOrderHandler(repo order.IOrderRepository) IOrder {
    return &orderHandler{repo: repo}
}
```

- `internal/router/` はハンドラの具象型ではなく、この `IXxx` を受け取る/保持する（`internal/router/route.go` の `route` 構造体を参照）

### ハンドラの依存の組み立ては `internal/router/route.go` に集約する 🔴

- リポジトリ・セッションマネージャ等の生成と、それを注入したハンドラの構築（`New*Handler` の呼び出し）は **`internal/router/route.go` の `router.New(root, cfg)`** で行う。`internal/server/server.go`（`Server.Run`）では行わない
- `Server.Run` は `router.New` が返す `IRouting`（`SetupApplicationRoute` / `SetupSystemRoute`）を呼ぶだけにする。サーバー起動処理（ミドルウェア登録・Listen）と依存の組み立てを分離するため
