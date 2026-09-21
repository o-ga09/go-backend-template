const yenFormatter = new Intl.NumberFormat('ja-JP', {
  style: 'currency',
  currency: 'JPY',
})

export function formatYen(priceYen: number): string {
  return yenFormatter.format(priceYen)
}
