import Link from 'next/link'
import { Card, CardContent, CardFooter, CardHeader } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { formatYen } from '@/lib/currency'
import type { Product } from '@/api/product/types'

export function ProductCard({ product }: { product: Product }) {
  return (
    <Link href={`/products/${product.id}`}>
      <Card className="h-full transition-colors hover:border-primary">
        <CardHeader>
          <h3 className="text-heading-4">{product.name}</h3>
        </CardHeader>
        <CardContent>
          <p className="text-body-sm text-muted-foreground line-clamp-2">{product.description}</p>
        </CardContent>
        <CardFooter className="flex items-center justify-between">
          <span className="text-body-lg font-semibold">{formatYen(product.priceYen)}</span>
          {product.stock === 0 && <Badge variant="secondary">在庫切れ</Badge>}
        </CardFooter>
      </Card>
    </Link>
  )
}
