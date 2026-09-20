# CLAUDE.md

このファイルは AI エージェント（Claude Code / その他 AGENTS.md 対応ツール）向けの、このリポジトリのエントリーポイント。`AGENTS.md` はこのファイルへのシンボリックリンク。

## プロジェクト概要

Go（Echo）+ Next.js（App Router）による Web アプリケーションテンプレート。認証込みの EC 商材サンプル実装を通じて、実運用を想定したレイヤー構成・エラーハンドリング・テスト方針を示すボイラープレート。詳細なセットアップ手順は [README.md](README.md)、アーキテクチャの設計判断は [DESIGN.md](DESIGN.md) を参照。

## 技術スタック

| 領域 | 技術 |
|---|---|
| Backend | Go 1.25 / Echo / GORM / MySQL |
| Frontend | Next.js（App Router）/ React / TypeScript / TanStack Query / pnpm |
| 認証 | NextAuth（Google OAuth） |
| インフラ | Cloud Run または ECS（`.github/workflows/deploy-*.yml`） |

## リポジトリ構成

```
backend/            Go API サーバー（cmd/, internal/, pkg/, db/）
frontend/           Next.js クライアント（app/, api/, components/, hooks/）
.claude/            ルール・スキル・hooks の正（.agents, .github/skills はここへのシンボリックリンク）
├── rules/          実装ルール（下記「詳細ルール」参照。会話に自動で読み込まれる）
└── skills/         タスク別の実行手順（下記「スキル」参照）
.github/
├── instructions/   GitHub Copilot 向けの同等の指示
└── workflows/      CI（テスト・Lint・セキュリティチェック・デプロイ）
```

## 開発コマンド

```bash
# Backend: MySQL + APIサーバー起動（ホットリロード）
cd backend && docker compose up --build

# Backend: マイグレーション
export DATABASE_URL="user:P@ssw0rd@tcp(localhost:3306)/test?parseTime=true"
go run cmd/migrate/main.go -command up

# Backend: テスト・Lint
go test ./... -coverprofile=coverage.out
golangci-lint run --config=./.golangci.yml ./...

# Frontend: 起動
cd frontend && pnpm install && pnpm dev

# Frontend: テスト・型チェック・Lint・フォーマット（コミット前に全て通すこと）
pnpm test
pnpm type-check
pnpm lint
pnpm format
```

## 詳細ルール（`.claude/rules/`）

実装時は必ず参照する。会話開始時に自動で読み込まれる。

| ファイル | 内容 |
|---|---|
| `architecture.md` | レイヤー構成（server → domain が基本形。usecase は複雑なオーケストレーション時のみ）、domain＝DBモデル、依存方向 |
| `package-structure.md` | ディレクトリ・パッケージ命名規則、ファイル分割の基準 |
| `error-handling.md` | `pkg/errors`（ergo ベース）の使い方、機密情報をエラーに載せない方針 |
| `context-propagation.md` | `context.Context` の伝搬方針、context に入れて良いもの／悪いもの |
| `transaction.md` | `ITransactionManager` の使い方、外部API呼び出しとトランザクションを重ねない方針 |
| `testing.md` | Go / フロントエンドのテスト方針（テーブル駆動テスト、モック方針） |
| `frontend.md` | Next.js のディレクトリ構成、データフェッチ（TanStack Query）、認証、禁止事項 |

**個別のプロダクト要件・ADR があればそちらを優先する。** ルールは実装の作法を決めるものであり、要件を上書きしない。

## スキル（`.claude/skills/`）

| スキル | 用途 |
|---|---|
| `implement-api` | `backend/` に新しい API エンドポイントを実装する |
| `implement-component` | `frontend/` に新しい画面・コンポーネントを実装する |
| `migration-db-schema` | `backend/db/migrations/` に DB マイグレーションを追加する |
| `design-feature` | 新機能の要件整理〜アーキテクチャ設計を行い、`DESIGN.md` に反映する |
| `commit` | 変更内容から Conventional Commits 形式のコミットメッセージを作成してコミットする |
| `create-pr` | 変更内容から Pull Request を作成する |
| `code-review` | PR またはブランチの差分をレビューする |

## CI（`.github/workflows/`）

- `lint_and_test.yml`: Backend のテスト・golangci-lint
- `security.yml`: gitleaks / trufflehog によるシークレット検出、SAST（Semgrep）、依存関係の脆弱性チェック（govulncheck / pnpm audit）
- `deploy-cloudrun.yml` / `deploy-ecs.yml`: デプロイ

## 設計ドキュメント

- [DESIGN.md](DESIGN.md): アーキテクチャ全体像と設計判断の理由
- `backend/docs/`: デプロイ手順（Cloud Run OIDC / ECS）
