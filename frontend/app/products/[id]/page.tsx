'use client'

import { useState } from 'react'
import { useParams } from 'next/navigation'
import { toast } from 'sonner'
import { useProduct } from '@/api/product/hooks'
import { useAddCartItem } from '@/api/cart/hooks'
import { useAuth } from '@/context/authContext'
import { formatYen } from '@/lib/currency'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import { ApiError } from '@/lib/apiClient'

export default function ProductDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { user, login } = useAuth()
  const { data: product, isLoading, isError } = useProduct(id)
  const addCartItem = useAddCartItem()
  const [quantity, setQuantity] = useState(1)

  if (isLoading) {
    return (
      <main className="container space-y-4 py-8">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-40 w-full max-w-xl" />
      </main>
    )
  }

  if (isError || !product) {
    return (
      <main className="container py-8">
        <p className="text-body-base text-destructive">商品が見つかりませんでした。</p>
      </main>
    )
  }

  const outOfStock = product.stock === 0

  const handleAddToCart = async () => {
    if (!user) {
      toast.error('カートに追加するにはログインが必要です')
      await login()
      return
    }
    try {
      await addCartItem.mutateAsync({ productId: product.id, quantity })
      toast.success('カートに追加しました')
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : 'カートへの追加に失敗しました')
    }
  }

  return (
    <main className="container max-w-2xl space-y-6 py-8">
      <div>
        <h1 className="text-heading-2">{product.name}</h1>
        {outOfStock && (
          <Badge variant="secondary" className="mt-2">
            在庫切れ
          </Badge>
        )}
      </div>

      <p className="text-body-base text-muted-foreground">{product.description}</p>
      <p className="text-heading-4">{formatYen(product.priceYen)}</p>
      <p className="text-body-sm text-muted-foreground">在庫: {product.stock}点</p>

      <div className="flex items-center gap-4">
        <Input
          type="number"
          min={1}
          max={Math.max(product.stock, 1)}
          value={quantity}
          disabled={outOfStock}
          onChange={e => {
            const next = Number(e.target.value)
            setQuantity(Number.isFinite(next) ? Math.min(Math.max(next, 1), product.stock) : 1)
          }}
          className="w-24"
        />
        <Button
          disabled={outOfStock || addCartItem.isPending}
          onClick={() => void handleAddToCart()}
        >
          {outOfStock ? '在庫切れ' : 'カートに追加'}
        </Button>
      </div>
    </main>
  )
}
