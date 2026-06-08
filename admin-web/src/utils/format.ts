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

export function centToYuan(cent: number) {
  return cent / 100
}

export function yuanToCent(yuan: number) {
  return Math.round(yuan * 100)
}
