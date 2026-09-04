import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import * as clinicsApi from '../api/clinics'
import { ApiError } from '../api/http'
import type { Clinic, ClinicCreateInput, ClinicUpdateInput, ProblemDetails } from '../types/api'
import { useDentistsStore } from './dentists'
import { usePaymentsStore } from './payments'

export const useClinicsStore = defineStore('clinics', () => {
  const items = ref(new Map<string, Clinic>())
  const loading = ref(false)
  const error = ref<ProblemDetails | null>(null)

  const list = computed(() =>
    [...items.value.values()].sort((a, b) => b.created_at.localeCompare(a.created_at)),
  )

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

  return { items, loading, error, list, create, fetch, update, remove }
})
