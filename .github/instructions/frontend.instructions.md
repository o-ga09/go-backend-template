---
applyTo: frontend/**
---

# Frontend Instructions (Next.js / React / TypeScript)

`general.instructions.md` の内容と重複しないよう、ここでは `frontend/` 配下のコード生成・変更時に守るべき詳細ルールへの導線と要点のみを示す。

## 詳細ルールの参照先

`frontend/` を変更するときは、`.claude/rules/frontend.md` に従う。矛盾する提案をしない。

## 生成時に必ず確認すること

- コンポーネントから直接 `fetch` しない。`api/<ドメイン>/` にクライアント関数 + TanStack Query フックを定義し、コンポーネントはフック経由で呼ぶ
- 新しいドメインの型・API クライアントは `api/<ドメイン名>/`（既存の `api/user/types.ts` の配置）に置く
- UI コンポーネントは shadcn/ui（`components/ui/`）+ Tailwind CSS を使う。新規コンポーネント追加は `npx shadcn@latest add <component>`
- フォーム入力は Zod スキーマ + React Hook Form（`@hookform/resolvers/zod`）で組む
- 認証状態の参照は `context/authContext.tsx` の `AuthProvider` 経由。コンポーネントで NextAuth の `useSession` を直接呼ぶ実装を増やさない
- OAuth クライアントシークレット等のサーバー専用値をクライアントコンポーネントに露出させない（`NEXT_PUBLIC_*` のみクライアントに渡せる）
- `any` を使わない
