import { apiFetch } from './http'
import type { Clinic, ClinicCreateInput, ClinicUpdateInput } from '../types/api'

export function createClinic(input: ClinicCreateInput): Promise<Clinic> {
  return apiFetch<Clinic>('/clinics', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function getClinic(id: string): Promise<Clinic> {
  return apiFetch<Clinic>(`/clinics/${id}`)
}

export function updateClinic(id: string, input: ClinicUpdateInput): Promise<Clinic> {
  return apiFetch<Clinic>(`/clinics/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  })
}

export function deleteClinic(id: string): Promise<void> {
  return apiFetch<void>(`/clinics/${id}`, { method: 'DELETE' })
}
