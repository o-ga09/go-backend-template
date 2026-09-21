'use client'

import Link from 'next/link'
import { toast } from 'sonner'
import { RequireAuth } from '@/components/require-auth'
import { useCart, useRemoveCartItem, useUpdateCartItem } from '@/api/cart/hooks'
import { useProducts } from '@/api/product/hooks'
import { formatYen } from '@/lib/currency'
import { ApiError } from '@/lib/apiClient'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { Separator } from '@/components/ui/separator'
import type { Product } from '@/api/product/types'

export default function CartPage() {
  return (
    <RequireAuth>
      <CartContent />
    </RequireAuth>
  )
}

function CartContent() {
  const { data: cart, isLoading: cartLoading, isError: cartError } = useCart()
  const { data: products, isLoading: productsLoading } = useProducts()
  const updateItem = useUpdateCartItem()
  const removeItem = useRemoveCartItem()

  const isLoading = cartLoading || productsLoading

  if (isLoading) {
    return (
      <main className="container max-w-2xl space-y-4 py-8">
        <Skeleton className="h-8 w-32" />
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
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

  const productById = new Map((products ?? []).map(p => [p.id, p]))
  const items = cart?.items ?? []

  if (items.length === 0) {
    return (
      <main className="container max-w-2xl space-y-4 py-8">
        <h1 className="text-heading-2">カート</h1>
        <p className="text-body-base text-muted-foreground">カートに商品がありません。</p>
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

  const handleQuantityChange = async (product: Product, quantity: number) => {
    try {
      await updateItem.mutateAsync({ productId: product.id, quantity })
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : '数量の変更に失敗しました')
    }
  }

  const handleRemove = async (productId: string) => {
    try {
      await removeItem.mutateAsync(productId)
      toast.success('カートから削除しました')
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : '削除に失敗しました')
    }
  }

  return (
    <main className="container max-w-2xl space-y-6 py-8">
      <h1 className="text-heading-2">カート</h1>

      <div className="space-y-4">
        {items.map(item => {
          const product = productById.get(item.productId)
          return (
            <div key={item.productId} className="flex items-center gap-4 rounded-lg border p-4">
              <div className="flex-1">
                <p className="text-body-base font-medium">
                  {product?.name ?? '商品情報を取得できません'}
                </p>
                {product && (
                  <p className="text-body-sm text-muted-foreground">
                    {formatYen(product.priceYen)}
                  </p>
                )}
              </div>
              <Input
                type="number"
                min={1}
                max={product?.stock ?? undefined}
                value={item.quantity}
                disabled={!product || updateItem.isPending}
                onChange={e => {
                  const next = Number(e.target.value)
                  if (product && Number.isFinite(next) && next >= 1) {
                    void handleQuantityChange(product, Math.min(next, product.stock))
                  }
                }}
                className="w-20"
              />
              <Button
                variant="ghost"
                disabled={removeItem.isPending}
                onClick={() => void handleRemove(item.productId)}
              >
                削除
              </Button>
            </div>
          )
        })}
      </div>

      <Separator />

      <div className="flex items-center justify-between">
        <span className="text-body-lg font-semibold">合計</span>
        <span className="text-heading-4">{formatYen(total)}</span>
      </div>

      <Button asChild className="w-full">
        <Link href="/checkout">レジに進む</Link>
      </Button>
    </main>
  )
}
