export const formatCurrency = (val) => {
  if (!val) return '0 ₽'
  const units = parseInt(val?.units || 0)
  const nano = parseInt(val?.nano || 0)
  const total = units + (nano / 1_000_000_000)
  return new Intl.NumberFormat('ru-RU', {
    minimumFractionDigits: 1,
    maximumFractionDigits: 1
  }).format(total) + ' ₽'
}

export const formatNumberAsCurrency = (num) => {
  return new Intl.NumberFormat('ru-RU', {
    minimumFractionDigits: 1,
    maximumFractionDigits: 1
  }).format(num) + ' ₽'
}

export const formatNumberAsPercent = (num) => {
  return new Intl.NumberFormat('ru-RU', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(num) + '%'
}