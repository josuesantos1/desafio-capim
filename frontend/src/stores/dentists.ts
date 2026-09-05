import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as dentistsApi from '../api/dentists'
import { ApiError } from '../api/http'
import type { Dentist, DentistCreateInput, DentistRolesInput, DentistUpdateInput, ProblemDetails } from '../types/api'
import { useClinicsStore } from './clinics'

export const useDentistsStore = defineStore('dentists', () => {
  const items = ref(new Map<string, Dentist>())
  const byClinic = ref(new Map<string, string[]>())
  const loading = ref(false)
  const error = ref<ProblemDetails | null>(null)

  function upsert(d: Dentist) {
    items.value.set(d.id, d)
    const ids = byClinic.value.get(d.clinic_id) ?? []
    if (!ids.includes(d.id)) {
      byClinic.value.set(d.clinic_id, [...ids, d.id])
    }
  }

  function listByClinic(clinicId: string): Dentist[] {
    const ids = byClinic.value.get(clinicId) ?? []
    return ids
      .map((id) => items.value.get(id))
      .filter((d): d is Dentist => d !== undefined)
  }

  async function create(clinicId: string, input: DentistCreateInput) {
    loading.value = true
    error.value = null
    try {
      const d = await dentistsApi.createDentist(clinicId, input)
      upsert(d)
      return d
    } catch (err) {
      if (err instanceof ApiError) error.value = err.problem
      throw err
    } finally {
      loading.value = false
    }
  }

  async function list(clinicId: string, limit?: number) {
    loading.value = true
    error.value = null
    try {
      const result = await dentistsApi.listDentists(clinicId, limit)
      byClinic.value.set(clinicId, [])
      for (const d of result.items) upsert(d)
      return result
    } catch (err) {
      if (err instanceof ApiError) error.value = err.problem
      throw err
    } finally {
      loading.value = false
    }
  }

  async function update(clinicId: string, id: string, input: DentistUpdateInput) {
    loading.value = true
    error.value = null
    try {
      const d = await dentistsApi.updateDentist(clinicId, id, input)
      upsert(d)
      return d
    } catch (err) {
      if (err instanceof ApiError) error.value = err.problem
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateRoles(clinicId: string, id: string, input: DentistRolesInput) {
    loading.value = true
    error.value = null
    try {
      const d = await dentistsApi.updateDentistRoles(clinicId, id, input)
      upsert(d)
      await useClinicsStore().fetch(clinicId)
      return d
    } catch (err) {
      if (err instanceof ApiError) error.value = err.problem
      throw err
    } finally {
      loading.value = false
    }
  }

  async function remove(clinicId: string, id: string) {
    loading.value = true
    error.value = null
    try {
      await dentistsApi.deleteDentist(clinicId, id)
      items.value.delete(id)
      const ids = byClinic.value.get(clinicId) ?? []
      byClinic.value.set(clinicId, ids.filter((i) => i !== id))
    } catch (err) {
      if (err instanceof ApiError) error.value = err.problem
      throw err
    } finally {
      loading.value = false
    }
  }

  function evictByClinicId(clinicId: string) {
    const ids = byClinic.value.get(clinicId) ?? []
    for (const id of ids) items.value.delete(id)
    byClinic.value.delete(clinicId)
  }

  return { items, byClinic, loading, error, listByClinic, create, list, update, updateRoles, remove, evictByClinicId }
})
