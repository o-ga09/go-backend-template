import { apiFetch } from '@/lib/apiClient'
import type { Cart, CartItem } from './types'

export function fetchCart(): Promise<Cart> {
  return apiFetch<Cart>('/api/cart')
}

export function addCartItem(item: CartItem): Promise<Cart> {
  return apiFetch<Cart>('/api/cart', { method: 'POST', body: item })
}

export function updateCartItem(item: CartItem): Promise<Cart> {
  return apiFetch<Cart>('/api/cart', { method: 'PUT', body: item })
}

export function removeCartItem(productId: string): Promise<Cart> {
  return apiFetch<Cart>('/api/cart', { method: 'DELETE', query: { productId } })
}

export function cartItemCount(cart: Cart | undefined): number {
  if (!cart) {
    return 0
  }
  return cart.items.reduce((sum, item) => sum + item.quantity, 0)
}
