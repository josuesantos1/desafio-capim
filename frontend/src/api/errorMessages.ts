import type { ProblemDetails } from '../types/api'

const overrides: Record<string, string> = {
  EMAIL_ALREADY_EXISTS: 'Este e-mail já está em uso por outro dentista.',
  DENTIST_NOT_FOUND: 'Dentista não encontrado — pode ter sido removido. Atualizando a lista.',
  INTERNAL_ERROR: 'Erro interno do servidor. Tente novamente.',
}

export function displayMessage(problem: ProblemDetails): string {
  if (problem.code && overrides[problem.code]) {
    return overrides[problem.code]
  }
  return problem.detail ?? problem.title
}
