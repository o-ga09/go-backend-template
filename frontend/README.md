This is a [Next.js](https://nextjs.org) project bootstrapped with [`create-next-app`](https://nextjs.org/docs/app/api-reference/cli/create-next-app).

## Getting Started

First, run the development server:

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
# or
bun dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

You can start editing the page by modifying `app/page.tsx`. The page auto-updates as you edit the file.

This project uses [`next/font`](https://nextjs.org/docs/app/building-your-application/optimizing/fonts) to automatically optimize and load [Geist](https://vercel.com/font), a new font family for Vercel.

## バックエンドAPIをモックして起動する(MSW)

`backend/` を起動せずに商品一覧・カート・注文系画面を確認したい場合、`NEXT_PUBLIC_API_MOCKING=enabled` を指定して起動すると [MSW](https://mswjs.io/) が `mocks/handlers.ts` 定義のレスポンスでAPIをモックする。

```bash
NEXT_PUBLIC_API_MOCKING=enabled pnpm dev
```

- モックの内容(レスポンス)を変えたい場合は `mocks/handlers.ts` を編集する(vitestのテストとブラウザの両方で同じハンドラを共有している)
- NextAuthのログインセッション自体はモックしていないため、Googleログインは引き続き必要(認証必須画面はログインしないと表示できない)
- 通常の `pnpm dev`(環境変数未指定)では従来通り実際のバックエンドAPIを呼び出す

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
