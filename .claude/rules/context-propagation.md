# Context 伝搬方針

対象: `backend/`（Go / Echo）。

## 基本ルール

- `context.Context` は**必ず第一引数**。変数名は `ctx`
- `context.Background()` を関数の途中で作らない。`main` のみで使う
- テストでは **`t.Context()`** を使う。`context.Background()` を test 内で作らない
- `context.TODO()` は使わない

```go
// 誤: 実装コードで context を自分で作る
func (s *sessionServer) FindByID(id string) error {
    ctx := context.Background()
    ...
}

// 正: Echo handler は c.Request().Context() を使う
func (s *sessionServer) FindByID(c echo.Context) error {
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

| キー定数 | 型 | Get 関数 | 格納タイミング |
|---|---|---|---|
| `config.ConfigKey` | `*config.Config` | `config.GetCtxEnv(ctx)` | 起動直後（`config.New`） |
| `ownerkey.OwnerKey` | `string` | `ownerkey.GetOwnerID(ctx)` | 認証ミドルウェア相当（RevenueCat anonymous App User ID） |

**新しい値を context に格納したい場合は `pkg/context/` に追加する。**
関数の引数で渡せるものは context に入れない。

MVP は認証を持たない（`ADR-0006`）。owner の識別は購入基盤の匿名IDをヘッダで受け取る簡易方式であり、
**これは認証ではない。** 権限判定の根拠に使わない。

## Context に入れてはいけないもの 🚫

- **復号した手紙本文・インタビュー回答・メモリ。** context の値はログ・トレースに載りやすく、`NFR-03-05`（平文をログに出さない）を破る最短経路になる。**平文は必ず引数で渡し、スコープを最小にする**
- 推しの名前
- ビジネスロジックのパラメータ全般

```go
// 誤: 復号した本文を context に載せる
ctx = context.WithValue(ctx, "answerText", plain)

// 正: 引数で渡す。使い終わったら参照を残さない
func (g *generator) Generate(ctx context.Context, materials []domain.Material) (*Draft, error)
```

## Get 関数の使い方

```go
// 正: pkg のヘルパーを使う
cfg := config.GetCtxEnv(ctx)

// 誤: ctx.Value を直接呼ぶ
cfg := ctx.Value("config").(*config.Config)
```

## キャンセルとタイムアウト

- Echo handler では `c.Request().Context()` をそのまま使う（Echo がキャンセルを管理）
- 外部呼び出しにはタイムアウトを付ける。`context.WithTimeout` は呼び出し元で行う

```go
// LLM 呼び出しは長い。NFR-01-04（下書き生成 30秒以内）に合わせる
callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()
draft, err := s.llm.GenerateDraft(callCtx, materials)
```

タイムアウト値の根拠は `docs/requirements.md` の NFR-01。

| 用途 | 目安 | 根拠 |
|---|---|---|
| 下書き生成 | 30秒 | NFR-01-04 |
| 「特にない」の意味的判定 | 3秒 | NFR-01-03a。超えたら掘り下げを諦めて次へ進み、ユーザーを待たせない |
| DB クエリ | 5秒 | — |

**LLM のタイムアウトでユーザーの回答を失わないこと**（FR-06-11 / NFR-02-03）。タイムアウト時も回答は永続化済みであること。

## Context の値を乱用しない

context はリクエストスコープの横断的関心事（Config, owner識別子）のためだけに使う。
ビジネスロジックのパラメータは引数で渡す。

```go
// 誤: ビジネスデータを context に入れる
ctx = context.WithValue(ctx, "oshiID", oshiID)

// 正: 引数で渡す
func (s *sessionServer) Start(ctx context.Context, oshiID string) (*Session, error)
```
