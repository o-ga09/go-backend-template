# リクエスト・バリデーションルール

対象: `backend/`（Go / Echo）。`internal/handler/request/` に置くリクエスト型と、ハンドラでのバインド・バリデーションの作法。

## リクエストのバインド 🔴

- リクエストは `internal/handler/request/` に定義した構造体に **`c.Bind(&req)` でバインドする**。`c.Param("id")` や `c.QueryParam("x")` をハンドラに直書きしない
- 構造体タグでパラメータの種類を使い分ける

| タグ | 用途 |
|---|---|
| `param` | パスパラメータ（例: `/users/:id` の `id`） |
| `query` | クエリパラメータ |
| `header` | リクエストヘッダー |
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

このプロジェクトが使う echo v5（`github.com/labstack/echo/v5`）の `DefaultBinder`（`engine.Binder` のデフォルト。明示的に差し替える必要はない）は `param`/`query`/`header` タグと、Content-Type に応じたボディ（`json` タグ等）を標準でバインドする。`Context` は v4 以前と異なり **interface ではなく struct** で、ハンドラ・ミドルウェアのシグネチャは `func(c *echo.Context) error` になる(`c echo.Context` ではなく `*echo.Context`)。

## バリデーション 🔴

- 構造体の形式的なバリデーション（必須・フォーマット等）は **`go-playground/validator/v10`** の `validate:"..."` タグで宣言する
- `c.Bind(&req)` の直後に **`c.Validate(&req)`** を呼ぶ。`engine.Validator` には `pkg/validator.New()` を登録している（`internal/server/server.go`）。ハンドラ・テストのどちらで `echo.New()` する場合も、`e.Validator = validator.New()` を設定すること
- バリデーションエラーは `errors.MakeBusinessError(ctx, err.Error())`（422）に変換する。`pkg/validator` はエラーメッセージ自体を組み立てて返すため、`err.Error()` をそのまま渡す（静的な固定文言で上書きしない）

```go
var req request.CreateUserRequest
if err := c.Bind(&req); err != nil {
    return errors.MakeBusinessError(ctx, "invalid request body")
}
if err := c.Validate(&req); err != nil {
    return errors.MakeBusinessError(ctx, err.Error())
}
```

### バリデーションエラーメッセージの日本語化（`ja` タグ）🔴

- クライアントに返すバリデーションエラーメッセージは **structタグ `ja` で宣言する**。`pkg/validator`（`go-playground/validator/v10` のラッパー）は、失敗したフィールドの `ja` タグの値をエラーメッセージとして使う
- `ja` タグが無いフィールドは `go-playground/validator` のデフォルト（英語）メッセージにフォールバックする。**クライアントへ返すメッセージは必ず `ja` タグで指定すること**（フォールバックはあくまで安全側の保険であり、意図して使うものではない）
- 複数フィールドが同時に違反した場合、`pkg/validator.ValidationError` は各メッセージを `、` で連結して返す

```go
type CreateUserRequest struct {
    UID         string `json:"uid" validate:"required" ja:"uidは必須です"`
    DisplayName string `json:"displayName" validate:"required" ja:"displayNameは必須です"`
    Email       string `json:"email,omitempty" validate:"omitempty,email" ja:"emailの形式が正しくありません"`
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
    UID          string `json:"uid" validate:"required" ja:"uidは必須です"`
    DisplayName  string `json:"displayName" validate:"required" ja:"displayNameは必須です"`
    Email        string `json:"email,omitempty" validate:"omitempty,email" ja:"emailの形式が正しくありません"` // 省略時はUIDから生成する
}
```
