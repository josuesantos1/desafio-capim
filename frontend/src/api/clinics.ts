import { apiFetch } from './http'
import type { Clinic, ClinicCreateInput, ClinicListParams, ClinicListResult, ClinicUpdateInput } from '../types/api'

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

export function listClinics(params: ClinicListParams = {}): Promise<ClinicListResult> {
  const query = new URLSearchParams()
  if (params.limit) query.set('limit', String(params.limit))
  if (params.offset) query.set('offset', String(params.offset))
  if (params.q) query.set('q', params.q)
  if (params.city) query.set('city', params.city)
  const qs = query.toString()
  return apiFetch<ClinicListResult>(`/clinics${qs ? `?${qs}` : ''}`)
}
