import type { AuthResponse, UserFullPortfolio } from "../types"

export async function fetchPortfolio(token: string): Promise<UserFullPortfolio> {
    const response = await fetch('/auth', {
        method: 'POST',
        headers: {
            'T-Token': token,
            'Content-Type': 'application/json',
        },
    })

    if (response.status === 401) {
        throw new Error('UNAUTHORIZED')
    }

    if (!response.ok) {
        throw new Error(`HTTP_ERROR_${response.status}`)
    }

    const data: AuthResponse = await response.json()
    return data.user_portfolio
}