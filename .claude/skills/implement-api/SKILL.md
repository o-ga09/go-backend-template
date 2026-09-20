---
name: implement-api
description: >
  backend/ に新しい API エンドポイントを実装するスキル。「API を追加して」「エンドポイントを実装して」
  「ハンドラーを追加して」と言われたときに使用する。.claude/rules/architecture.md のレイヤー方針
  （シンプルな CRUD はハンドラが domain を直接呼ぶ 2 層構成）に従って実装する。
---

# API 実装スキル

`backend/`（Go / Echo / GORM）に新しい API エンドポイントを実装する手順。

## 手順

1. **ドメインを用意する**（`internal/domain/<name>/entity.go`）
   - エンティティ、リポジトリ interface（`//go:generate moq -out mock/xxx_mock.go -pkg moq . IXxxRepository`）
   - バリデーション・状態遷移などのドメインロジックは純粋関数として実装する（ハンドラに書かない）
2. **リポジトリを実装する**（`internal/database/mysql/<name>.go`）
   - ビジネスロジックを持たせない。データの読み書きのみ
3. **リクエスト型を定義する**（`internal/handler/request/<name>.go`。`.claude/rules/request-validation.md` を参照）
   - パスパラメータは `param`、クエリパラメータは `query`、リクエストボディは `json` タグで宣言する
   - 形式的なバリデーション（必須・フォーマット等）は `go-validator/v10` の `validate:"..."` タグで宣言する
   - クライアントに返すバリデーションエラーメッセージは `ja` タグで日本語指定する（例: `validate:"required" ja:"uidは必須です"`）
   - フィールドごとに独立した doc コメントは書かない。必要な場合のみ行末にインラインコメントを書く
4. **ハンドラーを実装する**（`internal/handler/<name>.go`。`internal/handler/user.go` / `auth.go` を参照）
   - シンプルな CRUD はハンドラが domain のリポジトリを直接呼ぶ（usecase 層を作らない）
   - ハンドラ構造体は非公開（`xxxHandler`）にし、公開インターフェース `IXxxHandler` を定義して `NewXxxHandler` はそれを返す（`domain` のリポジトリ interface と同じ形。`package-structure.md`）
   - リクエストは `c.Param()`/`c.QueryParam()` を直書きせず、`c.Bind(&req)` → `c.Validate(&req)` の順で読み取る（ハンドラ・ミドルウェアのシグネチャは echo v5 に合わせ `func(c *echo.Context) error`）
   - ログインを必須とするエンドポイントは `Ctx.GetUserID(ctx)` が空なら `errors.MakeAuthorizedError`（401）で拒否し、対象リソース取得後に所有者チェック（domain の `IsOwnedBy` 等 + `pkg/authz.IsOwner`）で `errors.MakeAuthorizationError`（403）を返す（`architecture.md`「認可（所有者ベース）のパターン」）
   - 複数テーブルへの書き込みがある場合のみ `internal/database` の `ITransactionManager.RunInTx` でラップする
   - 複雑なオーケストレーション（複数リポジトリ・外部 API 呼び出しの組み合わせ）が必要な場合のみ `internal/service/` を挟む
5. **エラーハンドリング**
   - `pkg/errors` の `Make*Error` / `errors.Wrap` を使う。`fmt.Errorf` や標準 `errors.New` を直接使わない
   - **`c.Bind` / `c.Validate` の失敗は `errors.Wrap(ctx, err)` に渡すだけでよい。** ハンドラ側で `errors.MakeBusinessError` を呼ばない。422 コードは `pkg/validator`（`Validator.Validate` / `Binder.Bind`）が `errors.WithInvalidArgumentCode` で事前に付与しており、`Wrap` はそれを保ったままログ出力する（`request-validation.md`）
6. **ルーティング登録**（`internal/router/route.go` / `application.go`）
   - ハンドラの生成（リポジトリ・セッションマネージャ等の依存の組み立て）は `router.New(root, cfg)` に集約する。`internal/server/server.go` でハンドラを直接 `New*Handler` しない
   - `(*route).SetupApplicationRoute()` にエンドポイントを追加する
7. **モックを生成する**
   - `go generate ./...` で `//go:generate moq` ディレクティブからモックを再生成する
8. **テストを書く**（`.claude/rules/testing.md` を参照）
   - ドメイン層はテーブル駆動テストで純粋関数を検証する
   - ハンドラーは `httptest` + `echo.New()` で実際のルーターを立てて検証する。`e.Validator = validator.New()` と `e.Binder = validator.NewBinder()`（`internal/server/server.go` と同じ設定）を忘れない
   - リポジトリのテストは実 DB（`docker compose up -d db`）で検証する

## 参照

- `.claude/rules/architecture.md`
- `.claude/rules/package-structure.md`
- `.claude/rules/error-handling.md`
- `.claude/rules/context-propagation.md`
- `.claude/rules/transaction.md`
- `.claude/rules/request-validation.md`
- `.claude/rules/testing.md`
