import type { UserFullPortfolio } from "../types"

export async function fetchPortfolio(force: boolean, ctrl: AbortController): Promise <UserFullPortfolio> {
    const response = await fetch(`/api/portfolio?force=${force}`, {
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