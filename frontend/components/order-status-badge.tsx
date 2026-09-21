import { Badge } from '@/components/ui/badge'
import type { OrderStatus } from '@/api/order/types'

const STATUS_LABEL: Record<OrderStatus, string> = {
  pending: '処理中',
}

export function OrderStatusBadge({ status }: { status: OrderStatus }) {
  return <Badge variant="secondary">{STATUS_LABEL[status] ?? status}</Badge>
}
