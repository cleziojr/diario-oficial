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

export async function apiFetch(path: string, init?: RequestInit): Promise<Response> {
  const res = await fetch(`${apiPrefix()}${path}`, init)
  if (!res.ok) {
    throw new Error(await readErrorMessage(res))
  }
  return res
}
