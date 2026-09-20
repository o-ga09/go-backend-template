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
// ハンドラ側(internal/handler/user.go)
var req request.GetUserRequest
if err := c.Bind(&req); err != nil {
    return errors.Wrap(ctx, err)
}
```

このプロジェクトが使う echo v5（`github.com/labstack/echo/v5`）の `Context` は v4 以前と異なり **interface ではなく struct** で、ハンドラ・ミドルウェアのシグネチャは `func(c *echo.Context) error` になる(`c echo.Context` ではなく `*echo.Context`)。`engine.Binder` は `param`/`query`/`header`/`json` タグを標準でバインドする `echo.DefaultBinder` をラップした `pkg/validator.NewBinder()` を登録している（`internal/server/server.go`。次節参照）。

## バリデーション 🔴

- 構造体の形式的なバリデーション（必須・フォーマット等）は **`go-playground/validator/v10`** の `validate:"..."` タグで宣言する
- `c.Bind(&req)` の直後に **`c.Validate(&req)`** を呼ぶ。`engine.Validator` には `pkg/validator.New()` を登録している（`internal/server/server.go`）。ハンドラ・テストのどちらで `echo.New()` する場合も、`e.Validator = validator.New()` と `e.Binder = validator.NewBinder()` を設定すること

### `c.Bind` / `c.Validate` の失敗は `errors.Wrap(ctx, err)` するだけでよい 🔴

- **ハンドラ内で `c.Bind` / `c.Validate` の失敗を `errors.MakeBusinessError` に変換しない。** どちらも `errors.Wrap(ctx, err)` を呼ぶだけで自動的に 422 がレスポンスされる
- これは `pkg/validator` 側（`c.Validate` に登録する `Validator.Validate` と `c.Bind` に登録する `Binder.Bind`）が、失敗時に `errors.WithInvalidArgumentCode(err)` で **その場で 422 のエラーコードだけを事前に付与しておく**ことで成立する。`echo.Validator`/`echo.Binder` インターフェースのメソッドには `ctx` が渡らないため `errors.MakeBusinessError`（ログ出力にRequestIDが要る）は呼べないが、コードだけなら `ctx` 無しで付与できる
- コード付与済みの `err` は `ErrTypeBussiness` ではラップされていないため `errors.IsWrapped(err)` は `false` のまま。ハンドラが呼ぶ `errors.Wrap(ctx, err)` が実際にラップ・ログ出力を行い、その際 `ergo.CodeOf` が Unwrap チェーンをたどって既に付与済みの 422 コードを見つける（`error-handling.md`「エラーコードの付与」）
- 新しいバリデーション系のカスタムエラー（`c.Bind`/`c.Validate` 由来）を追加する場合も、この形（`pkg/validator` 側で `errors.WithInvalidArgumentCode` を付与し、ハンドラは `errors.Wrap` するだけ）に従う。ハンドラ側にバインド/バリデーション専用のラッパー関数を新設しない

```go
// pkg/validator/validator.go: echo.Validatorインターフェースの実装
func (v *Validator) Validate(i any) error {
    err := v.validate.Struct(i)
    if err == nil {
        return nil
    }
    verrs, ok := err.(val.ValidationErrors)
    if !ok {
        return err
    }
    return errors.WithInvalidArgumentCode(translate(i, verrs))
}

// pkg/validator/binder.go: echo.Binderインターフェースの実装。
// echo.DefaultBinderに委譲し、失敗した場合のみコードを付与する
func (b *Binder) Bind(c *echo.Context, target any) error {
    if err := b.delegate.Bind(c, target); err != nil {
        return errors.WithInvalidArgumentCode(err)
    }
    return nil
}
```

```go
// ハンドラ側(internal/handler/user.go)。誤ってここでMakeBusinessErrorを
// 呼ばない。c.Bind/c.Validateが返した時点で既に422確定のエラーなので、
// Wrapするだけで自動的に422がレスポンスされる
var req request.CreateUserRequest
if err := c.Bind(&req); err != nil {
    return errors.Wrap(ctx, err)
}
if err := c.Validate(&req); err != nil {
    return errors.Wrap(ctx, err)
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
