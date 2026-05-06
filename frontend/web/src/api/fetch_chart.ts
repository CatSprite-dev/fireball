import type { ChartData } from '../types'

export async function fetchChart(
    ctrl: AbortController,
    period: string = '1y',
    index: string = 'IMOEX'
): Promise<ChartData> {
    const response = await fetch(`/api/chart?period=${period}&index=${index}`, {
        method: 'GET',
        signal: ctrl.signal,
    })
    
    if (response.status === 401) {
        throw new Error('UNAUTHORIZED')
    }

    if (!response.ok) {
        throw new Error(`HTTP_ERROR_${response.status}`)
    }
    return response.json()
}