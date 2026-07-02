import { apiPrefix, authHeader, clearTokens, readErrorMessage, setTokens } from './client'

export type User = {
  id: string
  email: string
}

type TokenResponse = {
  access_token: string
  refresh_token: string
  user: User
}

async function postCredentials(path: string, email: string, password: string): Promise<User> {
  const res = await fetch(`${apiPrefix()}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  if (!res.ok) {
    throw new Error(await readErrorMessage(res))
  }
  const data = (await res.json()) as TokenResponse
  setTokens(data.access_token, data.refresh_token)
  return data.user
}

export function login(email: string, password: string): Promise<User> {
  return postCredentials('/api/v1/auth/login', email, password)
}

export function register(email: string, password: string): Promise<User> {
  return postCredentials('/api/v1/auth/register', email, password)
}

export async function getMe(): Promise<User> {
  const res = await fetch(`${apiPrefix()}/api/v1/auth/me`, { headers: authHeader() })
  if (!res.ok) {
    throw new Error('não autenticado')
  }
  return res.json() as Promise<User>
}

export function logout(): void {
  clearTokens()
}
