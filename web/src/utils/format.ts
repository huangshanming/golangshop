export function formatPrice(cents: number): string {
  return (cents / 100).toFixed(2)
}

export function formatPriceYuan(cents: number): { integer: string; decimal: string } {
  const yuan = (cents / 100).toFixed(2)
  const [integer, decimal] = yuan.split('.')
  return { integer, decimal }
}
