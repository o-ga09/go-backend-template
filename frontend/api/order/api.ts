import { apiFetch } from '@/lib/apiClient'
import type { Order } from './types'

export function fetchOrders(): Promise<Order[]> {
  return apiFetch<Order[]>('/api/orders')
}

export function fetchOrder(id: string): Promise<Order> {
  return apiFetch<Order>(`/api/orders/${id}`)
}

export function createOrder(): Promise<Order> {
  return apiFetch<Order>('/api/orders', { method: 'POST' })
}
