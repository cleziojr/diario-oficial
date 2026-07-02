export function apiPrefix(): string {
  const base = import.meta.env.VITE_API_BASE as string | undefined
  return (base ?? '').replace(/\/$/, '')
}

export async function readErrorMessage(res: Response): Promise<string> {
  try {
    const j: unknown = await res.json()
    if (
      j &&
      typeof j === 'object' &&
      'error' in j &&
      typeof (j as { error: unknown }).error === 'string'
    ) {
      return (j as { error: string }).error
    }
  } catch {
    /* ignore */
  }
  return res.statusText || `HTTP ${res.status}`
}

// --- Tokens (JWT access + refresh) -----------------------------------------

const ACCESS_KEY = 'access_token'
const REFRESH_KEY = 'refresh_token'

export function getAccessToken(): string | null {
  return localStorage.getItem(ACCESS_KEY)
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_KEY)
}

export function setTokens(access: string, refresh: string): void {
  localStorage.setItem(ACCESS_KEY, access)
  localStorage.setItem(REFRESH_KEY, refresh)
}

export function clearTokens(): void {
  localStorage.removeItem(ACCESS_KEY)
  localStorage.removeItem(REFRESH_KEY)
}

export function authHeader(): Record<string, string> {
  const t = getAccessToken()
  return t ? { Authorization: `Bearer ${t}` } : {}
}

// tryRefresh usa o refresh token para obter um novo par (rotação). Devolve
// true se conseguiu renovar.
export async function tryRefresh(): Promise<boolean> {
  const refresh = getRefreshToken()
  if (!refresh) return false
  try {
    const res = await fetch(`${apiPrefix()}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refresh }),
    })
    if (!res.ok) {
      clearTokens()
      return false
    }
    const data = (await res.json()) as { access_token: string; refresh_token: string }
    setTokens(data.access_token, data.refresh_token)
    return true
  } catch {
    return false
  }
}

// --- Fetch com auth + retry no 401 -----------------------------------------

function doFetch(path: string, init: RequestInit): Promise<Response> {
  const headers = { ...(init.headers ?? {}), ...authHeader() }
  return fetch(`${apiPrefix()}${path}`, { ...init, headers })
}

export async function apiFetch(path: string, init: RequestInit = {}): Promise<Response> {
  let res = await doFetch(path, init)
  if (res.status === 401 && (await tryRefresh())) {
    res = await doFetch(path, init)
  }
  if (!res.ok) {
    throw new Error(await readErrorMessage(res))
  }
  return res
}
