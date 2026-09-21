'use client'

import Link from 'next/link'
import { RequireAuth } from '@/components/require-auth'
import { useOrders } from '@/api/order/hooks'
import { useProducts } from '@/api/product/hooks'
import { formatYen } from '@/lib/currency'
import { OrderStatusBadge } from '@/components/order-status-badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Separator } from '@/components/ui/separator'

export default function OrderHistoryPage() {
  return (
    <RequireAuth>
      <OrderHistoryContent />
    </RequireAuth>
  )
}

function OrderHistoryContent() {
  const { data: orders, isLoading: ordersLoading, isError } = useOrders()
  const { data: products } = useProducts()
  const nameById = new Map((products ?? []).map(p => [p.id, p.name]))

  if (ordersLoading) {
    return (
      <main className="container max-w-2xl space-y-4 py-8">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-32 w-full" />
      </main>
    )
  }

  if (isError) {
    return (
      <main className="container py-8">
        <p className="text-body-base text-destructive">注文履歴の取得に失敗しました。</p>
      </main>
    )
  }

  if (!orders || orders.length === 0) {
    return (
      <main className="container max-w-2xl space-y-4 py-8">
        <h1 className="text-heading-2">注文履歴</h1>
        <p className="text-body-base text-muted-foreground">注文履歴はまだありません。</p>
        <Button asChild>
          <Link href="/">商品一覧を見る</Link>
        </Button>
      </main>
    )
  }

  return (
    <main className="container max-w-2xl space-y-6 py-8">
      <h1 className="text-heading-2">注文履歴</h1>

      <div className="space-y-4">
        {orders.map(order => (
          <div key={order.id} className="space-y-3 rounded-lg border p-4">
            <div className="flex items-center justify-between">
              <span className="text-body-sm text-muted-foreground">注文番号: {order.id}</span>
              <OrderStatusBadge status={order.status} />
            </div>
            <Separator />
            <div className="space-y-1">
              {order.items.map(item => (
                <div
                  key={item.productId}
                  className="flex items-center justify-between text-body-sm"
                >
                  <span>
                    {nameById.get(item.productId) ?? item.productId} × {item.quantity}
                  </span>
                  <span>{formatYen(item.unitPriceYen * item.quantity)}</span>
                </div>
              ))}
            </div>
            <Separator />
            <div className="flex items-center justify-between font-semibold">
              <span>合計</span>
              <span>{formatYen(order.totalPriceYen)}</span>
            </div>
          </div>
        ))}
      </div>
    </main>
  )
}
