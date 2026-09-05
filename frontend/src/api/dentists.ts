import { apiFetch } from './http'
import type {
  Dentist,
  DentistCreateInput,
  DentistListResult,
  DentistRolesInput,
  DentistUpdateInput,
} from '../types/api'

export function createDentist(clinicId: string, input: DentistCreateInput): Promise<Dentist> {
  return apiFetch<Dentist>(`/clinics/${clinicId}/dentists`, {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function listDentists(clinicId: string, limit?: number): Promise<DentistListResult> {
  const query = limit ? `?limit=${limit}` : ''
  return apiFetch<DentistListResult>(`/clinics/${clinicId}/dentists${query}`)
}

export function updateDentist(
  clinicId: string,
  id: string,
  input: DentistUpdateInput,
): Promise<Dentist> {
  return apiFetch<Dentist>(`/clinics/${clinicId}/dentists/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  })
}

export function updateDentistRoles(
  clinicId: string,
  id: string,
  input: DentistRolesInput,
): Promise<Dentist> {
  return apiFetch<Dentist>(`/clinics/${clinicId}/dentists/${id}/roles`, {
    method: 'PATCH',
    body: JSON.stringify(input),
  })
}

export function deleteDentist(clinicId: string, id: string): Promise<void> {
  return apiFetch<void>(`/clinics/${clinicId}/dentists/${id}`, { method: 'DELETE' })
}
