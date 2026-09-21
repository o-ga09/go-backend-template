# CLAUDE.md

このファイルは AI エージェント（Claude Code / その他 AGENTS.md 対応ツール）向けの、このリポジトリのエントリーポイント。`AGENTS.md` はこのファイルへのシンボリックリンク。

## 応答言語 🔴

**ユーザーへの回答・コメント・コミットメッセージ・PR本文などの生成は日本語で行う。** コード自体（識別子・コメント内の技術用語等）やコマンド・ログ出力はこの限りではない。

## プロジェクト概要

Go（Echo）+ Next.js（App Router）による Web アプリケーションテンプレート。認証込みの EC 商材サンプル実装を通じて、実運用を想定したレイヤー構成・エラーハンドリング・テスト方針を示すボイラープレート。詳細なセットアップ手順は [README.md](README.md)、UIデザイン（デザインシステム・画面仕様）は [DESIGN.md](DESIGN.md) を参照。バックエンドのアーキテクチャ設計判断は本ファイルの「アーキテクチャの設計判断」を参照。

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
| `request-validation.md` | リクエスト型のバインド（`c.Bind` + `param`/`query`/`json` タグ）・`go-validator/v10` によるバリデーション（`ja` タグでエラーメッセージを日本語化）、リクエスト型のコメント方針 |
| `testing.md` | Go / フロントエンドのテスト方針（テーブル駆動テスト、モック方針） |
| `frontend.md` | Next.js のディレクトリ構成、データフェッチ（TanStack Query）、認証、禁止事項 |

**個別のプロダクト要件・ADR があればそちらを優先する。** ルールは実装の作法を決めるものであり、要件を上書きしない。

## アーキテクチャの設計判断

`.claude/rules/` が「どう書くか」を定めるのに対し、ここでは「なぜそう決めたか」を簡潔にまとめる。

- **レイヤーは2層がデフォルト（`server` → `domain`）。** usecase 層（`internal/service/`）は、複数リポジトリをまたぐオーケストレーション・外部API呼び出し・暗号化を伴う場合のみの例外。シンプルな CRUD に usecase 層を強制すると、3層すべてが「呼ぶだけ」の薄いパススルーになりやすく、複雑さに見合わない
- **domain＝DBモデル。** domain の構造体がそのまま GORM モデルとして永続化される前提にすることで、ドメイン層とデータアクセス層の変換コードを不要にする。`BaseModel` を埋め込めば採番・タイムスタンプ・楽観ロックを GORM プラグインに委譲できる。暗号化やカラム名の詰め替えが必要なドメインだけこの前提から外れる
- **外部API呼び出しとトランザクションを重ねない。** 応答時間が読めない外部呼び出しの間 DB 接続を占有すると、Cloud Run のスケール0起動時などに接続が枯渇しやすいため、「外部API呼び出し（トランザクション外）→ 短いトランザクションでの書き込み」の順序に固定する
- **エラーは `pkg/errors`（ergoベース）に統一する。** エラーコードとスタックトレースを一貫した形で持たせ、`errors.GetCode(err)` から HTTP ステータスへ機械的に変換できるようにする。エラーメッセージに機密情報を載せない運用を、レビューで検出しやすい形（`Make*Error` の使用箇所を見るだけで判断できる）にする
- **機密情報は `context.Context` に入れない。** パスワード・トークン・個人情報・決済情報は引数で渡し、生存スコープを最小にする
- **在庫のような競合しうるカウンタの排他制御は、楽観ロック（`Version`）ではなく条件付き `UPDATE ... WHERE stock >= ?` で行う。**（Issue #4, `internal/database/mysql/product.go` の `DecreaseStock`）在庫チェック（`HasStock`）から実際の減算までの間に他の注文が割り込む TOCTOU を、条件付きUPDATE自体が防ぐ。`UpdateColumn` + `gorm.Expr` を使うことで `BaseModelPlugin` の楽観ロックフック（`beforeUpdate`）を意図的にバイパスしており（渡す構造体の `Version` がゼロ値のため早期returnする）、この操作では `Version`/`UpdatedAt` は更新されない。事前チェック後の競合（`RowsAffected == 0`）は 409（`errors.MakeConflictError`）として扱う
- **Cookie認証下の状態変更エンドポイントには CSRF 対策を必須とする。**（Issue #4, `internal/server/middleware.go` の `CSRFProtection`）セッションCookieは別オリジンのフロントエンドから `credentials: 'include'` で送るため `SameSite=None` で発行しており、これはクロスサイトリクエストでも自動送信される。echo v5組み込みのCSRFミドルウェアを使い、`Sec-Fetch-Site` ヘッダー（Fetch Metadata）による検証を優先することで、フロントエンド側の実装変更なしに正規のクロスオリジンリクエストのみを許可する（詳細は `backend/docs/auth.md`「CSRF対策」）

新しいドメイン（EC商材の商品・カート・注文など）を追加する場合も、上記の方針（2層デフォルト・domain＝DBモデル・パッケージ構成）に従う。個別機能の設計は `design-feature` スキルで整理し、既存方針からの逸脱を含む大きな設計判断はこのセクションに追記する。

## スキル（`.claude/skills/`）

| スキル | 用途 |
|---|---|
| `implement-api` | `backend/` に新しい API エンドポイントを実装する |
| `implement-component` | `frontend/` に新しい画面・コンポーネントを実装する |
| `migration-db-schema` | `backend/db/migrations/` に DB マイグレーションを追加する |
| `design-feature` | 新機能の要件整理〜アーキテクチャ設計を行い、UIに関わる設計は `DESIGN.md`、大きなアーキテクチャ判断は本ファイルに反映する |
| `commit` | 変更内容から Conventional Commits 形式のコミットメッセージを作成してコミットする |
| `create-pr` | 変更内容から Pull Request を作成する |
| `code-review` | PR またはブランチの差分をレビューする |

## CI（`.github/workflows/`）

- `lint_and_test.yml`: Backend のテスト・golangci-lint
- `security.yml`: gitleaks / trufflehog によるシークレット検出、SAST（Semgrep）、依存関係の脆弱性チェック（govulncheck / pnpm audit）
- `deploy-cloudrun.yml` / `deploy-ecs.yml`: デプロイ

## 設計ドキュメント

- [DESIGN.md](DESIGN.md): `frontend/` の UI デザイン（デザインシステム・EC商材サンプルの画面仕様）
- 本ファイルの「アーキテクチャの設計判断」: バックエンドの設計判断の理由
- `backend/docs/`: デプロイ手順（Cloud Run OIDC / ECS）
