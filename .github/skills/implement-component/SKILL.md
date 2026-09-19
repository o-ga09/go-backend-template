---
name: implement-component
description: >
  frontend/ に新しい画面・コンポーネントを実装するスキル。「画面を追加して」「コンポーネントを実装して」
  「ページを作って」と言われたときに使用する。.claude/rules/frontend.md の TanStack Query・shadcn/ui・
  Zod の方針に従って実装する。
---

# コンポーネント実装スキル

`frontend/`（Next.js / React / TypeScript）に新しい画面・コンポーネントを実装する手順。

## 手順

1. **API クライアント・型を用意する**（`api/<domain>/`）
   - `api/<domain>/types.ts` にレスポンス/リクエストの型を定義する（`any` を使わない）
   - TanStack Query のクエリ/ミューテーションフックを同ディレクトリに定義する
2. **バリデーションが必要な場合は Zod スキーマを定義する**
   - フォームは React Hook Form + `@hookform/resolvers/zod` を使う
3. **画面・コンポーネントを実装する**
   - ページは `app/` 配下（App Router の規約に従う）、再利用可能な部品は `components/`
   - 1 ファイル = 1 コンポーネント
   - UI は shadcn/ui（`components/ui/`）+ Tailwind CSS を使う。新規 shadcn コンポーネントは `npx shadcn@latest add <component>` で追加する
   - データ取得はコンポーネントから直接 `fetch` せず、手順 1 のフックを使う
   - 認証状態が必要な画面は `context/authContext.tsx` の `AuthProvider` から取得する
4. **テストを書く**（`.claude/rules/testing.md` を参照）
   - ロジック（フック・ユーティリティ）は Vitest で検証する
5. **確認する**
   - `pnpm lint` / `pnpm type-check` / `pnpm test` / `pnpm format` が通ることを確認する

## 参照

- `.claude/rules/frontend.md`
- `.claude/rules/testing.md`
