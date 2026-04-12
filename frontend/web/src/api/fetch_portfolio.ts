import type { UserFullPortfolio } from "../types"

export async function fetchPortfolio(): Promise <UserFullPortfolio> {
    const response = await fetch('/api/portfolio', {
        method: 'POST',
    })

    if (response.status === 401) {
        throw new Error('UNAUTHORIZED')
    }

    if (!response.ok) {
        throw new Error(`HTTP_ERROR_${response.status}`)
    }

    return response.json()
}