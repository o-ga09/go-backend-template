'use client'

import { useRouter, useSearchParams } from 'next/navigation'
import { useProducts } from '@/api/product/hooks'
import { filterProductsByName } from '@/api/product/api'
import { ProductCard } from '@/components/product-card'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'

export function ProductList() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const query = searchParams.get('q') ?? ''
  const { data: products, isLoading, isError } = useProducts()

  const handleQueryChange = (value: string) => {
    const params = new URLSearchParams(searchParams.toString())
    if (value) {
      params.set('q', value)
    } else {
      params.delete('q')
    }
    router.replace(`/?${params.toString()}`)
  }

  const filtered = products ? filterProductsByName(products, query) : []

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-heading-2">商品一覧</h1>
        <Input
          className="mt-4 max-w-sm"
          placeholder="商品名で検索"
          defaultValue={query}
          onChange={e => handleQueryChange(e.target.value)}
        />
      </div>

      {isLoading && (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-56 w-full" />
          ))}
        </div>
      )}

      {isError && (
        <p className="text-body-base text-destructive">
          商品一覧の取得に失敗しました。時間をおいて再度お試しください。
        </p>
      )}

      {!isLoading && !isError && filtered.length === 0 && (
        <p className="text-body-base text-muted-foreground">該当する商品が見つかりませんでした。</p>
      )}

      {!isLoading && !isError && filtered.length > 0 && (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {filtered.map(product => (
            <ProductCard key={product.id} product={product} />
          ))}
        </div>
      )}
    </div>
  )
}
