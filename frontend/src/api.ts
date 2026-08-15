export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers)
  if (options.body && !(options.body instanceof FormData) && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')
  const response = await fetch(`/api/v1${path}`, { ...options, headers })
  if (!response.ok) {
    const body = await response.json().catch(() => null)
    throw new Error(body?.error?.message ?? `Erro HTTP ${response.status}`)
  }
	if (response.status === 204) return undefined as T
	const content = await response.text()
	return (content ? JSON.parse(content) : undefined) as T
}

export function json(method: string, body: unknown): RequestInit { return { method, body: JSON.stringify(body) } }
export const items = <T,>(value: { items?: T[] }) => value.items ?? []
