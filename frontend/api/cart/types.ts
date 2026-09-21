export type CartItem = {
  productId: string
  quantity: number
}

export type Cart = {
  id: string
  items: CartItem[]
}
