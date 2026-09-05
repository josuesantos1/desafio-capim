import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import * as clinicsApi from '../api/clinics'
import { ApiError } from '../api/http'
import type {
  Clinic,
  ClinicCreateInput,
  ClinicListParams,
  ClinicUpdateInput,
  ProblemDetails,
} from '../types/api'
import { useDentistsStore } from './dentists'
import { usePaymentsStore } from './payments'

export const useClinicsStore = defineStore('clinics', () => {
  const items = ref(new Map<string, Clinic>())
  const loading = ref(false)
  const error = ref<ProblemDetails | null>(null)

  const lastResult = ref<{
    status: 'success' | 'error'
    ids: string[]
    total: number
    limit: number
    offset: number
  } | null>(null)

  async function create(input: ClinicCreateInput) {
    loading.value = true
    error.value = null
    try {
      const c = await clinicsApi.createClinic(input)
      items.value.set(c.id, c)
      return c
    } catch (err) {
      if (err instanceof ApiError) error.value = err.problem
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetch(id: string) {
    loading.value = true
    error.value = null
    try {
      const c = await clinicsApi.getClinic(id)
      items.value.set(c.id, c)
      return c
    } catch (err) {
      if (err instanceof ApiError) error.value = err.problem
      throw err
    } finally {
      loading.value = false
    }
  }

  async function update(id: string, input: ClinicUpdateInput) {
    loading.value = true
    error.value = null
    try {
      const c = await clinicsApi.updateClinic(id, input)
      items.value.set(c.id, c)
      return c
    } catch (err) {
      if (err instanceof ApiError) error.value = err.problem
      throw err
    } finally {
      loading.value = false
    }
  }

  async function remove(id: string) {
    loading.value = true
    error.value = null
    try {
      await clinicsApi.deleteClinic(id)
      items.value.delete(id)
      useDentistsStore().evictByClinicId(id)
      usePaymentsStore().evictByClinicId(id)
    } catch (err) {
      if (err instanceof ApiError) error.value = err.problem
      throw err
    } finally {
      loading.value = false
    }
  }

  async function list(params: ClinicListParams = {}) {
    loading.value = true
    error.value = null
    try {
      const result = await clinicsApi.listClinics(params)
      for (const c of result.items) items.value.set(c.id, c)
      lastResult.value = {
        status: 'success',
        ids: result.items.map((c) => c.id),
        total: result.total,
        limit: result.limit,
        offset: result.offset,
      }
      return result
    } catch (err) {
      if (err instanceof ApiError) error.value = err.problem
      lastResult.value = {
        status: 'error',
        ids: [],
        total: 0,
        limit: params.limit ?? 0,
        offset: params.offset ?? 0,
      }
      throw err
    } finally {
      loading.value = false
    }
  }

  const results = computed(() =>
    (lastResult.value?.ids ?? [])
      .map((id) => items.value.get(id))
      .filter((c): c is Clinic => c !== undefined),
  )

  const resultsStatus = computed(() => lastResult.value?.status ?? null)

  // AVISO: checagem 100% cosmética/client-side — o backend não valida posse
  // nenhuma, qualquer chamada direta à API ignora isso. Ver spec-mock-login.md
  // (Overview) para o risco aceito conscientemente.
  function isOwner(clinic: Clinic, email: string | null): boolean {
    return email !== null && clinic.email === email
  }

  return {
    items,
    loading,
    error,
    create,
    fetch,
    update,
    remove,
    list,
    results,
    resultsStatus,
    isOwner,
  }
})
