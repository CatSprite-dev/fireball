import type { AuthResponse, UserFullPortfolio } from "../types"

export async function fetchPortfolio(): Promise<UserFullPortfolio> {
    const response = await fetch('/auth', {
        method: 'POST',
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

export async function login(token: string) {
    const response = await fetch('/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ token })
    })

    if (response.status === 401) {
        throw new Error('UNAUTHORIZED')
    }

    if (!response.ok) {
        throw new Error(`HTTP_ERROR_${response.status}`)
    }

    return 
}