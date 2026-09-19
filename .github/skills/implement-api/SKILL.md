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
3. **ハンドラーを実装する**（`internal/handler/<name>.go`）
   - シンプルな CRUD はハンドラが domain のリポジトリを直接呼ぶ（usecase 層を作らない）
   - 複数テーブルへの書き込みがある場合のみ `internal/infra/database/transaction.ITransactionManager.RunInTx` でラップする
   - 複雑なオーケストレーション（複数リポジトリ・外部 API 呼び出しの組み合わせ）が必要な場合のみ `internal/service/` を挟む
4. **エラーハンドリング**
   - `pkg/errors` の `Make*Error` / `errors.Wrap` を使う。`fmt.Errorf` や標準 `errors.New` を直接使わない
5. **ルーティング登録**（`internal/router/application.go`）
   - `SetupApplicationRoute` にエンドポイントを追加する
6. **モックを生成する**
   - `go generate ./...` で `//go:generate moq` ディレクティブからモックを再生成する
7. **テストを書く**（`.claude/rules/testing.md` を参照）
   - ドメイン層はテーブル駆動テストで純粋関数を検証する
   - ハンドラーは `httptest` + `echo.New()` で実際のルーターを立てて検証する
   - リポジトリのテストは実 DB（`docker compose up -d db`）で検証する

## 参照

- `.claude/rules/architecture.md`
- `.claude/rules/package-structure.md`
- `.claude/rules/error-handling.md`
- `.claude/rules/context-propagation.md`
- `.claude/rules/transaction.md`
- `.claude/rules/testing.md`
