'use client'

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { cartKeys } from '@/api/cart/hooks'
import { createOrder, fetchOrder, fetchOrders } from './api'

export const orderKeys = {
  all: ['orders'] as const,
  detail: (id: string) => ['orders', id] as const,
}

export function useOrders() {
  return useQuery({
    queryKey: orderKeys.all,
    queryFn: fetchOrders,
  })
}

export function useOrder(id: string) {
  return useQuery({
    queryKey: orderKeys.detail(id),
    queryFn: () => fetchOrder(id),
    enabled: id.length > 0,
  })
}

export function useCreateOrder() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createOrder,
    onSuccess: () => {
      // 注文確定でカートがクリアされるため、カート・注文履歴のキャッシュを破棄する
      void queryClient.invalidateQueries({ queryKey: cartKeys.all })
      void queryClient.invalidateQueries({ queryKey: orderKeys.all })
    },
  })
}
