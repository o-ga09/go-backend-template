---
applyTo: *
---

# Copilot Instructions

このリポジトリはGo(Echo)とNext.jsを使用したWebアプリケーションテンプレートです。

## アーキテクチャ概要

```
backend/                    # Go API サーバー (Echo framework)
├── cmd/api/main.go         # エントリーポイント
├── cmd/migrate/main.go     # DBマイグレーションツール
├── internal/
│   ├── domain/             # ドメインモデル・ビジネスロジック
│   ├── handler/            # HTTPハンドラー (Controller層)
│   ├── service/            # アプリケーションサービス
│   ├── infra/database/     # DB接続・インフラ層
│   ├── router/             # ルーティング定義
│   └── server/             # HTTPサーバー・ミドルウェア
├── pkg/                    # 共有パッケージ (config, logger, errors, context)
frontend/                   # Next.js App Router
├── app/                    # ページ・APIルート
├── api/                    # APIレスポンス型定義
├── providers/              # React Context プロバイダー
└── components/ui/          # shadcn/ui コンポーネント
```

## 開発ワークフロー

### 起動コマンド

```bash
# Docker Compose でバックエンド・DB起動 (ホットリロード対応)
cd backend && docker compose up

# フロントエンド開発サーバー
cd frontend && pnpm dev
```

### バックエンドコマンド

```bash
# マイグレーション実行
go run cmd/migrate/main.go -command up

# 新規マイグレーション作成
go run cmd/migrate/main.go -command new -name create_users

# シード投入
go run cmd/migrate/main.go -command seed
```

### フロントエンドコマンド

```bash
pnpm lint          # oxlint 型認識モード
pnpm test          # vitest
pnpm type-check    # tsgo (TypeScriptネイティブコンパイラ)
pnpm format:fix    # prettier
```

## 重要なパターン

### Context経由のDI (Backend)

設定、DB接続、RequestIDは`context.Context`経由で伝播:

```go
// pkg/context/context.go のヘルパーを使用
cfg := Ctx.GetCfgFromCtx(ctx)
db := Ctx.GetDBFromCtx(ctx)
requestID := Ctx.GetRequestID(ctx)
```

### エラーハンドリング (Backend)

`pkg/errors`パッケージの`Make*Error`関数でエラー分類・ロギングを統一:

```go
errors.MakeBusinessError(ctx, "不正な操作です")    // 400系
errors.MakeNotFoundError(ctx, "ユーザーが見つかりません")
errors.MakeSystemError(ctx, err)                  // 500系
```

### ルート追加 (Backend)

新規APIは`internal/router/application.go`の`SetupApplicationRoute`に追加:

```go
func SetupApplicationRoute(root *echo.Group) {
    users := root.Group("/users")
    users.GET("", handler.ListUsers)
}
```

### 認証フロー (Frontend)

NextAuth(Google OAuth) → `AuthContext` → バックエンドAPI:

- [app/api/auth/[...nextauth]/routes.ts](frontend/app/api/auth/[...nextauth]/routes.ts) - NextAuth設定
- [context/authContext.tsx](frontend/context/authContext.tsx) - 認証状態管理
- [providers/apiProvider.tsx](frontend/providers/apiProvider.tsx) - TanStack Query設定

### APIデータ取得 (Frontend)

TanStack Queryを使用。型定義は`api/`ディレクトリに配置:

```typescript
// api/user/types.ts に型定義
// hooks/ にカスタムフック
```

## 環境変数

### Backend

- `PORT` - サーバーポート (default: 80)
- `DATABASE_URL` - MySQL接続文字列
- `PROJECTID` - GCP Project ID (Cloud Logging用)
- `ENV` - 環境名 (dev/prod)

### Frontend

- `NEXT_PUBLIC_API_BASE_URL` - バックエンドAPIのURL
- `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` - OAuth認証
- `NEXTAUTH_SECRET` - セッション暗号化キー

## UIライブラリ

- [shadcn/ui](https://ui.shadcn.com/) + Radix UI + Tailwind CSS
- コンポーネント追加: `npx shadcn@latest add <component>`
