'use client'

import { useState } from 'react'
import Link from 'next/link'
import { toast } from 'sonner'
import { RequireAuth } from '@/components/require-auth'
import { useCart } from '@/api/cart/hooks'
import { useCreateOrder } from '@/api/order/hooks'
import { useProducts } from '@/api/product/hooks'
import { formatYen } from '@/lib/currency'
import { ApiError } from '@/lib/apiClient'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Separator } from '@/components/ui/separator'
import type { Order } from '@/api/order/types'

export default function CheckoutPage() {
  return (
    <RequireAuth>
      <CheckoutContent />
    </RequireAuth>
  )
}

function CheckoutContent() {
  const { data: cart, isLoading: cartLoading, isError: cartError } = useCart()
  const { data: products, isLoading: productsLoading } = useProducts()
  const createOrder = useCreateOrder()
  const [completedOrder, setCompletedOrder] = useState<Order | null>(null)

  if (completedOrder) {
    return <CheckoutComplete order={completedOrder} products={products ?? []} />
  }

  if (cartLoading || productsLoading) {
    return (
      <main className="container max-w-xl space-y-4 py-8">
        <Skeleton className="h-8 w-32" />
        <Skeleton className="h-40 w-full" />
      </main>
    )
  }

  if (cartError) {
    return (
      <main className="container py-8">
        <p className="text-body-base text-destructive">カートの取得に失敗しました。</p>
      </main>
    )
  }

  const items = cart?.items ?? []
  const productById = new Map((products ?? []).map(p => [p.id, p]))

  if (items.length === 0) {
    return (
      <main className="container max-w-xl space-y-4 py-8">
        <h1 className="text-heading-2">注文の確認</h1>
        <p className="text-body-base text-muted-foreground">カートが空です。</p>
        <Button asChild>
          <Link href="/">商品一覧を見る</Link>
        </Button>
      </main>
    )
  }

  const total = items.reduce((sum, item) => {
    const product = productById.get(item.productId)
    return sum + (product ? product.priceYen * item.quantity : 0)
  }, 0)

  const handleSubmit = async () => {
    try {
      const order = await createOrder.mutateAsync()
      setCompletedOrder(order)
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : '注文の確定に失敗しました')
    }
  }

  return (
    <main className="container max-w-xl space-y-6 py-8">
      <h1 className="text-heading-2">注文の確認</h1>

      <div className="space-y-3 rounded-lg border p-4">
        {items.map(item => {
          const product = productById.get(item.productId)
          return (
            <div key={item.productId} className="flex items-center justify-between">
              <span className="text-body-base">
                {product?.name ?? '商品情報を取得できません'} × {item.quantity}
              </span>
              {product && (
                <span className="text-body-base">
                  {formatYen(product.priceYen * item.quantity)}
                </span>
              )}
            </div>
          )
        })}
        <Separator />
        <div className="flex items-center justify-between font-semibold">
          <span>合計</span>
          <span>{formatYen(total)}</span>
        </div>
      </div>

      <p className="text-body-sm text-muted-foreground">
        本テンプレートでは配送先・支払い方法の入力は未実装です。上記内容で注文を確定します。
      </p>

      <Button
        className="w-full"
        disabled={createOrder.isPending}
        onClick={() => void handleSubmit()}
      >
        {createOrder.isPending ? '注文を確定しています…' : '注文を確定する'}
      </Button>
    </main>
  )
}

function CheckoutComplete({
  order,
  products,
}: {
  order: Order
  products: { id: string; name: string }[]
}) {
  const nameById = new Map(products.map(p => [p.id, p.name]))

  return (
    <main className="container max-w-xl space-y-6 py-8">
      <h1 className="text-heading-2">ご注文ありがとうございました</h1>
      <p className="text-body-base text-muted-foreground">注文番号: {order.id}</p>

      <div className="space-y-3 rounded-lg border p-4">
        {order.items.map(item => (
          <div key={item.productId} className="flex items-center justify-between">
            <span className="text-body-base">
              {nameById.get(item.productId) ?? item.productId} × {item.quantity}
            </span>
            <span className="text-body-base">{formatYen(item.unitPriceYen * item.quantity)}</span>
          </div>
        ))}
        <Separator />
        <div className="flex items-center justify-between font-semibold">
          <span>合計</span>
          <span>{formatYen(order.totalPriceYen)}</span>
        </div>
      </div>

      <div className="flex gap-4">
        <Button asChild variant="outline" className="flex-1">
          <Link href="/">商品一覧に戻る</Link>
        </Button>
        <Button asChild className="flex-1">
          <Link href="/mypage/orders">注文履歴を見る</Link>
        </Button>
      </div>
    </main>
  )
}
