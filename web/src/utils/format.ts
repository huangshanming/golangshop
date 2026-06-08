export function formatPrice(cents: number): string {
  return (cents / 100).toFixed(2)
}

export function formatPriceYuan(cents: number): { integer: string; decimal: string } {
  const yuan = (cents / 100).toFixed(2)
  const [integer, decimal] = yuan.split('.')
  return { integer, decimal }
}

/** 统一格式：YYYY-MM-DD HH:mm:ss */
export function formatDateTime(value: string | number | Date | null | undefined): string {
  if (value == null || value === '') return '-'
  if (typeof value === 'string' && /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/.test(value)) {
    return value
  }
  let date: Date
  if (value instanceof Date) {
    date = value
  } else if (typeof value === 'number') {
    date = new Date(value > 1e12 ? value : value * 1000)
  } else {
    date = new Date(value)
  }
  if (Number.isNaN(date.getTime())) return String(value)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}
