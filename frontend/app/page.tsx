import { Suspense } from 'react'
import { ProductList } from '@/components/product-list'
import { Skeleton } from '@/components/ui/skeleton'

export default function Home() {
  return (
    <main className="container py-8">
      <Suspense fallback={<Skeleton className="h-64 w-full" />}>
        <ProductList />
      </Suspense>
    </main>
  )
}
