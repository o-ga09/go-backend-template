'use client'

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { addCartItem, fetchCart, removeCartItem, updateCartItem } from './api'
import type { Cart, CartItem } from './types'

export const cartKeys = {
  all: ['cart'] as const,
}

export function useCart(enabled = true) {
  return useQuery({
    queryKey: cartKeys.all,
    queryFn: fetchCart,
    enabled,
  })
}

function useCartMutation(mutationFn: (item: CartItem) => Promise<Cart>) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn,
    onSuccess: cart => {
      queryClient.setQueryData(cartKeys.all, cart)
    },
  })
}

export function useAddCartItem() {
  return useCartMutation(addCartItem)
}

export function useUpdateCartItem() {
  return useCartMutation(updateCartItem)
}

export function useRemoveCartItem() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (productId: string) => removeCartItem(productId),
    onSuccess: cart => {
      queryClient.setQueryData(cartKeys.all, cart)
    },
  })
}
