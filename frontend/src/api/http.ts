import type { ProblemDetails } from '../types/api'

export class ApiError extends Error {
  problem: ProblemDetails

  constructor(problem: ProblemDetails) {
    super(problem.detail ?? problem.title)
    this.problem = problem
  }
}

function isProblemDetails(body: unknown): body is ProblemDetails {
  return (
    typeof body === 'object' &&
    body !== null &&
    'status' in body &&
    'title' in body
  )
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`/api${path}`, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...init?.headers },
  })

  if (res.status === 204) {
    return undefined as T
  }

  const body: unknown = await res.json().catch(() => null)

  if (!res.ok) {
    if (isProblemDetails(body)) {
      throw new ApiError(body)
    }
    throw new ApiError({ type: 'about:blank', title: 'Unexpected error', status: res.status })
  }

  return body as T
}
