import type { Product } from '@/api/product/types'
import type { Cart } from '@/api/cart/types'
import type { Order } from '@/api/order/types'

/**
 * ブラウザでのモック起動(NEXT_PUBLIC_API_MOCKING=enabled)時に画面へ表示する
 * サンプルデータ。テスト側は各テストでserver.use()により個別に上書きするため、
 * ここでの内容はvitestの結果には影響しない。
 */
export const mockProducts: Product[] = [
  {
    id: 'p1',
    name: 'コーヒー豆(ブレンド) 200g',
    description: '酸味と苦味のバランスが取れた定番ブレンド。中煎り。',
    priceYen: 1200,
    stock: 20,
  },
  {
    id: 'p2',
    name: '紅茶(アールグレイ) 50g',
    description: 'ベルガモットの香り豊かなアールグレイ。',
    priceYen: 900,
    stock: 15,
  },
  {
    id: 'p3',
    name: 'マグカップ',
    description: 'シンプルな白磁のマグカップ。容量350ml。',
    priceYen: 1500,
    stock: 0,
  },
]

export const mockCart: Cart = {
  id: 'cart-1',
  items: [{ productId: mockProducts[0].id, quantity: 2 }],
}

export const mockOrders: Order[] = [
  {
    id: 'order-1',
    status: 'pending',
    totalPriceYen: mockProducts[0].priceYen,
    items: [{ productId: mockProducts[0].id, quantity: 1, unitPriceYen: mockProducts[0].priceYen }],
  },
]
