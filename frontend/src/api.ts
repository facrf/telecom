export class ApiError extends Error {
  status: number
  code?: string
  field?: string

  constructor(message: string, status: number, code?: string, field?: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.field = field
  }
}

export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers)
  if (options.body && !(options.body instanceof FormData) && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')
  const response = await fetch(`/api/v1${path}`, { ...options, headers })
  if (!response.ok) {
    const body = await response.json().catch(() => null)
    throw new ApiError(body?.error?.message ?? `Erro HTTP ${response.status}`, response.status, body?.error?.code, body?.error?.field)
  }
	if (response.status === 204) return undefined as T
	const content = await response.text()
	return (content ? JSON.parse(content) : undefined) as T
}

export function json(method: string, body: unknown): RequestInit { return { method, body: JSON.stringify(body) } }
export const items = <T,>(value: { items?: T[] }) => value.items ?? []
export function localDateTime(date = new Date()) {
  const offset = date.getTimezoneOffset() * 60_000
  return new Date(date.getTime() - offset).toISOString().slice(0, 16)
}
