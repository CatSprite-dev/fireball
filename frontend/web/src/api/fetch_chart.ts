import type { UserFullPortfolio, ChartData } from '../types'

export async function fetchChart(
    portfolio: UserFullPortfolio,
    period: string = '1y',
    index: string = 'IMOEX'
): Promise<ChartData> {
    const response = await fetch(`/api/chart?period=${period}&index=${index}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(portfolio),
    })
    
    if (response.status === 401) {
        throw new Error('UNAUTHORIZED')
    }

    if (!response.ok) {
        throw new Error(`HTTP_ERROR_${response.status}`)
    }
    return response.json()
}