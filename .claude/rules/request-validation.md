# リクエスト・バリデーションルール

対象: `backend/`（Go / Echo）。`internal/handler/request/` に置くリクエスト型と、ハンドラでのバインド・バリデーションの作法。

## リクエストのバインド 🔴

- リクエストは `internal/handler/request/` に定義した構造体に **`c.Bind(&req)` でバインドする**。`c.Param("id")` や `c.QueryParam("x")` をハンドラに直書きしない
- 構造体タグでパラメータの種類を使い分ける

| タグ | 用途 |
|---|---|
| `param` | パスパラメータ（例: `/users/:id` の `id`） |
| `query` | クエリパラメータ |
| `json` | リクエストボディ（JSON） |

```go
// internal/handler/request/user.go
type GetUserRequest struct {
    ID string `param:"id" validate:"required"`
}

type ListUsersRequest struct {
    Limit int `query:"limit" validate:"omitempty,max=100"`
}

type CreateUserRequest struct {
    DisplayName string `json:"displayName" validate:"required"`
}
```

```go
// ハンドラ側
var req request.GetUserRequest
if err := c.Bind(&req); err != nil {
    return errors.MakeBusinessError(ctx, "invalid request")
}
```

このプロジェクトが使う echo v3（`labstack/echo` v3.3.10）の `DefaultBinder` は `param` タグを解釈しない（echo v4 にはあるが v3 にはない）。そのため `engine.Binder` には `pkg/binder.New()`（`param` タグのバインドを追加で行い、`query`/`json` は echo 標準の `DefaultBinder` に委譲するラッパー）を登録している（`internal/server/server.go`）。**ハンドラ・テストのどちらで `echo.New()` する場合も、`e.Binder = binder.New()` を設定すること。** 設定を忘れると `param` タグのフィールドが空のままバインドされ、`validate:"required"` で 422 になる。

## バリデーション 🔴

- 構造体の形式的なバリデーション（必須・フォーマット等）は **`go-playground/validator/v10`** の `validate:"..."` タグで宣言する
- `c.Bind(&req)` の直後に **`c.Validate(&req)`** を呼ぶ。`engine.Validator` には `pkg/validator.New()` を登録している（`internal/server/server.go`）。ハンドラ・テストのどちらで `echo.New()` する場合も、`e.Validator = validator.New()` を設定すること
- バリデーションエラーは `errors.MakeBusinessError(ctx, msg)`（422）に変換する

```go
var req request.CreateUserRequest
if err := c.Bind(&req); err != nil {
    return errors.MakeBusinessError(ctx, "invalid request body")
}
if err := c.Validate(&req); err != nil {
    return errors.MakeBusinessError(ctx, "invalid request body")
}
```

### validate タグとドメイン層のバリデーションの役割分担

- `validate` タグは**形式的な検証**（必須・型・フォーマット等）に使う
- **ビジネスルールの判定（状態遷移の可否など）は `architecture.md`「ドメイン層のルール」の通り domain 層の純粋関数に置く。** `validate` タグで置き換えない
- 両方が同じ入力を検証することはある（例: 必須チェックを `validate:"required"` と domain 層の両方で行う）。これは重複ではなく、責務が異なる（形式 vs. ビジネスルール）ため許容する

## リクエスト型のコメント方針 🔴

- **フィールドごとに独立した doc コメント（フィールドの直前行にコメント行を書く形）を書かない。** 構造体・パッケージ単位のコメントは通常通り書いてよい
- フィールドの意図が名前・タグだけで伝わらない場合は、**その行の末尾にインラインコメントとして**書く

```go
// 誤: フィールドごとに独立したdocコメント
type CreateUserRequest struct {
    // UID はGoogleのsub。
    UID string `json:"uid"`
    // DisplayName はユーザーの表示名。
    DisplayName string `json:"displayName"`
}

// 正: 自明なフィールドにはコメントを書かない。必要な場合のみインラインで
type CreateUserRequest struct {
    UID          string `json:"uid" validate:"required"`
    DisplayName  string `json:"displayName" validate:"required"`
    Email        string `json:"email,omitempty" validate:"omitempty,email"` // 省略時はUIDから生成する
}
```
