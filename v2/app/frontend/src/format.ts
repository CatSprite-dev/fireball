// Показ денег и процентов. Числа приходят из Go строкой; в число — только для показа, считать с ним нельзя.

const money = new Intl.NumberFormat('ru-RU', {minimumFractionDigits: 2, maximumFractionDigits: 2})
const signed = new Intl.NumberFormat('ru-RU', {minimumFractionDigits: 2, maximumFractionDigits: 2, signDisplay: 'exceptZero'})

// Минус — типографский, как в макете.
function minus(s: string): string {
  return s.replace('-', '−')
}

// «1 044 651,48»; пустая строка — прочерк.
export function rub(s: string | undefined): string {
  return s ? minus(money.format(Number(s))) : '—'
}

// «+10 421,47» / «−2 231,90» / «0,00».
export function signedRub(s: string | undefined): string {
  return s ? minus(signed.format(Number(s))) : '—'
}

// «+15,50%»; unit — «%» или « п.п.».
export function signedPercent(s: string | undefined, unit = '%'): string {
  return s ? `${minus(signed.format(Number(s)))}${unit}` : `—${unit}`
}

// Класс цвета по знаку: pos, neg или muted.
export function tone(s: string | undefined): string {
  const n = Number(s)
  if (!s || n === 0) return 'muted'
  return n > 0 ? 'pos' : 'neg'
}

// «2026-11-21» → «21.11.2026».
export function day(s: string | undefined): string {
  if (!s) return '—'
  const [y, m, d] = s.split('-')
  return `${d}.${m}.${y}`
}
