# フロントエンドルール（Next.js）

対象: `frontend/`（Next.js / React / TypeScript）。バックエンドは `architecture.md` ほかを参照。
**このプロジェクトのクライアントは Next.js（App Router）1つだけ。** React Native / Expo のようなモバイルアプリは存在しない。

## 技術スタック

Next.js（App Router）+ React + TypeScript + **TanStack Query**。パッケージマネージャは **pnpm**。

| 用途 | コマンド |
|---|---|
| 起動 | `pnpm dev` |
| ビルド | `pnpm build` / `pnpm start` |
| 型チェック | `pnpm type-check`（**tsgo**。標準の `tsc` ではない） |
| Lint | `pnpm lint`（**oxlint** 型認識モード。ESLint ではない） |
| テスト | `pnpm test`（**Vitest**） |
| フォーマット | `pnpm format`（チェックのみ） / `pnpm format:fix` |

コミット前に `pnpm lint` / `pnpm type-check` / `pnpm test` / `pnpm format` が通ること。

## ディレクトリ構成

```
frontend/
├── app/            App Router（ページ・APIルート。認証コールバック等）
├── api/            APIレスポンス/リクエストの型定義（ドメインごと）
├── components/
│   └── ui/         shadcn/ui コンポーネント
├── context/        React Context（例: authContext）
├── providers/      Provider（QueryClientProvider・NextAuth SessionProvider）
├── hooks/          カスタムフック
├── lib/            汎用ユーティリティ（フォント・メタデータ・ローダー設定等）
├── public/         静的アセット
└── tests/          テストのセットアップ（`vitest.config.ts` の `setupFiles`）
```

新しいドメインの型・APIクライアント・フックは `api/<ドメイン名>/` にまとめる（既存の `api/user/types.ts` の配置に合わせる）。

## データフェッチ

- サーバー状態は **TanStack Query** で管理する（`providers/apiProvider.tsx` の `QueryClientProvider` でラップ済み）。`useState` + `useEffect` で自前にフェッチしない
- コンポーネントから直接 `fetch` を呼ばない。`api/<ドメイン>/` にクライアント関数 + クエリ/ミューテーションフックを定義し、コンポーネントはフックだけを使う
- API のベース URL は環境変数 `NEXT_PUBLIC_API_BASE_URL` を参照する

```tsx
// 正: ドメインごとのフック経由
const { data, isLoading } = useUser()

// 誤: コンポーネントから直接フェッチ
useEffect(() => {
  fetch(`${baseURL}/api/users`)
}, [])
```

## 認証（NextAuth）

- 認証基盤は **NextAuth**（Google OAuth）。設定は `app/api/auth/[...nextauth]/`
- 認証状態は `context/authContext.tsx`（`AuthProvider`）で管理し、コンポーネントはこの Context 経由でユーザー情報を参照する
- NextAuth のセッション確立後、バックエンド側のユーザー識別・作成 API（`GET /api/auth/user` 等）を呼び出して自前のユーザー情報を同期する。バックエンド API のエンドポイント設計は `architecture.md` に従う
- **NextAuth のクライアントシークレット（`GOOGLE_CLIENT_SECRET`, `NEXTAUTH_SECRET` 等）はサーバーサイド（`app/api/`）でのみ扱い、クライアントコンポーネントに露出させない**

## バリデーション

- 入力バリデーションは **Zod** スキーマで定義する
- フォームは **React Hook Form** + `@hookform/resolvers/zod` を使う
- API レスポンスの型は `api/<ドメイン>/types.ts` に定義し、`any` を使わない

## コンポーネント設計

- 1 ファイル = 1 コンポーネント
- UI コンポーネントは **shadcn/ui**（Radix UI + Tailwind CSS）を使う。追加は `npx shadcn@latest add <component>`
- スタイリングは Tailwind CSS を基本とする

## 型

- `any` を使わない
- API から返る ID・列挙値などはユニオン型/リテラル型にし、素の `string` を安易に使わない

## テスト

テストの詳細（対象範囲・書き方）は `testing.md` を参照する。

## 禁止事項 🚫

- コンポーネントから直接 `fetch` / HTTP クライアントを呼ぶ
- サーバー専用のシークレット（OAuth クライアントシークレット、`NEXTAUTH_SECRET` 等）をクライアントコンポーネントや `NEXT_PUBLIC_*` 以外の想定で露出させる
- 状態管理ライブラリ（Redux 等）の追加導入（TanStack Query + Context で足りる）
- `utils/`, `helpers/`, `common/` のような曖昧なディレクトリ名を作る
