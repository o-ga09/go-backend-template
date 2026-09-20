# Context 伝搬方針

対象: `backend/`（Go / Echo）。

## 基本ルール

- `context.Context` は**必ず第一引数**。変数名は `ctx`
- `context.Background()` を関数の途中で作らない。`main` のみで使う
- テストでは **`t.Context()`** を使う。`context.Background()` を test 内で作らない
- `context.TODO()` は使わない

```go
// 誤: 実装コードで context を自分で作る
func (s *orderServer) FindByID(id string) error {
    ctx := context.Background()
    ...
}

// 正: Echo handler は c.Request().Context() を使う
func (s *orderServer) FindByID(c echo.Context) error {
    ctx := c.Request().Context()
    ...
}

// 正: テストは t.Context() を使う
func TestFoo(t *testing.T) {
    ctx := t.Context()
    ...
}
```

## Context に格納する値

このプロジェクトで context に格納する値は `pkg/context/` で一元管理する。

| キー定数 | 型 | Get / Set 関数 | 格納タイミング |
|---|---|---|---|
| `config.ConfigKey` | `*config.Config` | `context.GetCfgFromCtx(ctx)` | 起動直後（`config.New`） |
| `RequestIDKey` | `string` | `context.GetRequestID` / `SetRequestID` | `AddID` ミドルウェア |
| `RequestTimeKey` | `time.Time` | `context.GetRequestTime` / `SetRequestTime` | `AddTime` ミドルウェア |
| `DBKey` | `*gorm.DB` | `context.GetDBFromCtx` / `SetDB` | `SetDB` ミドルウェア |
| `UserIDKey` | `string` | `context.GetUserID` / `SetUserID` | `Authenticate` ミドルウェア（セッションCookieが有効な場合のみ。`architecture.md`「認可（所有者ベース）のパターン」） |

**新しい値を context に格納したい場合は `pkg/context/` に追加する。**
関数の引数で渡せるものは context に入れない。

認証（ログインユーザーの識別子など）を実装する際も、識別子そのものは context 値として `pkg/context/` に追加してよいが、**それを権限判定の根拠として直接使う場合は認可ミドルウェア／ドメイン層で明示的に検証する。** context に入っていることを「認証済みである証明」として扱わない。

現在の実装例（`UserIDKey`）: `Authenticate` ミドルウェアはセッションCookieが無い/不正でもリクエストを拒否せずそのまま次へ進める（`context.GetUserID(ctx)` が空文字になるだけ）。**ログインを必須にするかどうかの判定は必ずハンドラ側で `GetUserID(ctx) == ""` を見て行う。** ミドルウェアが「認証済み」を保証しているわけではない。

## Context に入れてはいけないもの 🚫

- **パスワード・トークンなどの機密情報、個人情報（メールアドレス・氏名等）の平文。** context の値はログ・トレースに載りやすく、うっかり出力してしまう最短経路になる。**これらは必ず引数で渡し、生存スコープを最小にする**
- ビジネスロジックのパラメータ全般（ドメインID、リクエストパラメータなど）

```go
// 誤: 機密情報を context に載せる
ctx = context.WithValue(ctx, "plainPassword", pw)

// 正: 引数で渡す。使い終わったら参照を残さない
func (s *authService) Login(ctx context.Context, email string, password secret.Text) (*Session, error)
```

## Get 関数の使い方

```go
// 正: pkg のヘルパーを使う
cfg := context.GetCfgFromCtx(ctx)

// 誤: ctx.Value を直接呼ぶ
cfg := ctx.Value("config").(*config.Config)
```

## キャンセルとタイムアウト

- Echo handler では `c.Request().Context()` をそのまま使う（Echo がキャンセルを管理）
- 全リクエスト共通のタイムアウトは `WithTimeout` ミドルウェアが設定する
- それより長い/短い独自のタイムアウトが必要な外部呼び出し（決済APIなど）は、呼び出し元で個別に `context.WithTimeout` を設定する

```go
// 外部APIの応答が遅い可能性がある呼び出しには個別にタイムアウトを設定する
callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
defer cancel()
res, err := s.paymentClient.Charge(callCtx, req)
```

タイムアウト値は呼び出し先の SLA や要件に応じてドメイン/機能ごとに決める。根拠なく値だけ揃えない。

## Context の値を乱用しない

context はリクエストスコープの横断的関心事（Config、リクエストID、DBハンドルなど）のためだけに使う。
ビジネスロジックのパラメータは引数で渡す。

```go
// 誤: ビジネスデータを context に入れる
ctx = context.WithValue(ctx, "orderID", orderID)

// 正: 引数で渡す
func (s *orderServer) Create(ctx context.Context, orderID string) (*Order, error)
```
