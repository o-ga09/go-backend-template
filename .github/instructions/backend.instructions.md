---
applyTo: backend/**
---

# Backend Instructions (Go / Echo / GORM)

`general.instructions.md` の内容と重複しないよう、ここでは `backend/` 配下のコード生成・変更時に守るべき詳細ルールへの導線と要点のみを示す。

## 詳細ルールの参照先

`backend/` を変更するときは、以下のルールファイルに必ず従う。矛盾する提案をしない。

- `.claude/rules/architecture.md` — レイヤー構成（シンプルな CRUD はハンドラが domain を直接呼ぶ 2 層構成。複雑なオーケストレーションのみ `internal/service/` を挟む）
- `.claude/rules/package-structure.md` — ディレクトリ・パッケージ命名規則、interface の置き場所
- `.claude/rules/error-handling.md` — `pkg/errors`（ergo ベース）経由のエラー処理。`fmt.Errorf` / 標準 `errors.New` を直接使わない
- `.claude/rules/context-propagation.md` — `context.Context` は必ず第一引数。`context.Background()` は `main` のみ
- `.claude/rules/transaction.md` — 複数テーブル書き込みは `ITransactionManager.RunInTx` でラップする
- `.claude/rules/request-validation.md` — リクエストは `c.Bind` + `param`/`query`/`header`/`json` タグでバインドし、`go-validator/v10` の `validate` タグでバリデーションする。エラーメッセージは `ja` タグで日本語指定する。リクエスト型にフィールドごとの doc コメントを書かない
- `.claude/rules/testing.md` — テーブル駆動テスト、`moq` によるモック、実 DB でのリポジトリテスト

## 生成時に必ず確認すること

- 新規ドメインは `internal/domain/<name>/entity.go` にエンティティ + `//go:generate moq` の interface を定義する
- 新規 API は `internal/router/application.go`（または `system.go`）に登録する
- エラーは `pkg/errors.Make*Error` / `errors.Wrap` を経由する（`logger.Warn` との重複ログを作らない）
- マイグレーションは `db/migrations/`（`sql-migrate` 形式）で管理する。GORM の `AutoMigrate` は使わない
- 環境変数は `pkg/config.Config` に追加し、`pkg/context.GetCfgFromCtx(ctx)` で取得する
