export function useFormatters() {
    function formatCurrency(num: number): string {
        return new Intl.NumberFormat('ru-RU', {
            style: 'currency',
            currency: 'RUB',
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
        }).format(num)
    }

    return { formatCurrency }
}

export function parseMoney(m: { units: string; nano: number } | undefined): number {
    if (!m) return 0
    return (parseFloat(m.units || '0') || 0) + (m.nano || 0) / 1e9
}