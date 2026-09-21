import { apiFetch } from '@/lib/apiClient'
import type { Product } from './types'

export function fetchProducts(): Promise<Product[]> {
  return apiFetch<Product[]>('/api/products')
}

export function fetchProduct(id: string): Promise<Product> {
  return apiFetch<Product>(`/api/products/${id}`)
}

/**
 * 商品名の部分一致で商品一覧を絞り込む(バックエンドは検索クエリ未対応のため
 * クライアント側でフィルタする)。大文字・小文字は区別しない。
 */
export function filterProductsByName(products: Product[], query: string): Product[] {
  const normalized = query.trim().toLowerCase()
  if (!normalized) {
    return products
  }
  return products.filter(product => product.name.toLowerCase().includes(normalized))
}
