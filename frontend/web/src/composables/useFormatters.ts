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