---
name: migration-db-schema
description: >
  backend/db/migrations/ に DB マイグレーションを追加するスキル。「マイグレーションを作って」
  「テーブルを追加して」「スキーマを変更して」と言われたときに使用する。GORM の AutoMigrate は
  使わず、sql-migrate（cmd/migrate）で管理する方針に従う。
---

# DB マイグレーションスキル

`backend/`（MySQL / GORM / `sql-migrate`）のスキーマ変更を行う手順。

## 前提

- スキーマの正は `db/migrations/` の SQL ファイル。**GORM の `AutoMigrate` は使わない**
- マイグレーションの実行・管理は `cmd/migrate/main.go`（`sql-migrate` ラッパー）で行う
- 実行には `DATABASE_URL` 環境変数が必要（`docker compose up -d db` でローカル MySQL を起動する）

## 手順

1. **新規マイグレーションファイルを作成する**

   ```bash
   cd backend
   go run cmd/migrate/main.go -command new -name <migration_name>
   ```

   `db/migrations/<timestamp>_<migration_name>.sql` が生成される（`-- +migrate Up` / `-- +migrate Down` の雛形入り）。

2. **SQL を記述する**
   - `-- +migrate Up` に変更後のスキーマ、`-- +migrate Down` にロールバック用の SQL を書く
   - 新規ドメインのテーブルには `.claude/rules/architecture.md` の `BaseModel`（`id` / `version` / `created_at` / `updated_at`）に対応するカラムを含める
   - 監査項目（作成者等）が必要な場合はテーブルごとに `created_user_id` 等を追加する

3. **マイグレーションを適用して確認する**

   ```bash
   go run cmd/migrate/main.go -command up
   go run cmd/migrate/main.go -command status
   ```

4. **ロールバックできることを確認する**

   ```bash
   go run cmd/migrate/main.go -command down
   ```

5. **開発用シードが必要な場合**
   - `db/seed/*.sql` にシード SQL を追加する
   - `go run cmd/migrate/main.go -command seed` で投入されることを確認する

6. **GORM モデル側を追従させる**
   - `internal/domain/<name>/entity.go` のエンティティ・GORM タグを新しいスキーマに合わせる（domain＝DB モデル方針。`architecture.md`）
   - リポジトリ（`internal/database/mysql/<name>.go`）のクエリを確認する

## 参照

- `.claude/rules/architecture.md`（`BaseModel`・楽観ロックの節）
- `.claude/rules/transaction.md`
