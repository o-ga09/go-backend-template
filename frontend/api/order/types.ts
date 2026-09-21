export type OrderItem = {
  productId: string
  quantity: number
  unitPriceYen: number
}

export type OrderStatus = 'pending'

export type Order = {
  id: string
  status: OrderStatus
  totalPriceYen: number
  items: OrderItem[]
}
