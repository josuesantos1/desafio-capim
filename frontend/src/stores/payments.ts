import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as paymentsApi from '../api/payments'
import { ApiError } from '../api/http'
import type { Payment, PaymentCreateInput, PaymentListParams, ProblemDetails } from '../types/api'
import { useDentistsStore } from './dentists'
import { mockSettlementDate } from '../mock/finance'

export const usePaymentsStore = defineStore('payments', () => {
  const items = ref(new Map<string, Payment>())
  const byClinic = ref(new Map<string, string[]>())
  const loading = ref(false)
  const error = ref<ProblemDetails | null>(null)

  function upsert(p: Payment) {
    items.value.set(p.id, p)
    const ids = byClinic.value.get(p.clinic_id) ?? []
    if (!ids.includes(p.id)) {
      byClinic.value.set(p.clinic_id, [p.id, ...ids])
    }
  }

  function listByClinic(clinicId: string): Payment[] {
    const ids = byClinic.value.get(clinicId) ?? []
    return ids
      .map((id) => items.value.get(id))
      .filter((p): p is Payment => p !== undefined)
  }

  async function create(input: PaymentCreateInput) {
    loading.value = true
    error.value = null
    try {
      const idempotencyKey = crypto.randomUUID()
      const p = await paymentsApi.createPayment(input, idempotencyKey)
      upsert(p)
      return p
    } catch (err) {
      if (err instanceof ApiError) {
        error.value = err.problem
        if (err.problem.code === 'DENTIST_NOT_FOUND') {
          await useDentistsStore().list(input.clinic_id)
        }
      }
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchOne(id: string) {
    loading.value = true
    error.value = null
    try {
      const p = await paymentsApi.getPayment(id)
      upsert(p)
      return p
    } catch (err) {
      if (err instanceof ApiError) error.value = err.problem
      throw err
    } finally {
      loading.value = false
    }
  }

  async function list(clinicId: string, params: PaymentListParams = {}) {
    loading.value = true
    error.value = null
    try {
      const result = await paymentsApi.listPayments(clinicId, params)
      for (const p of result.items) items.value.set(p.id, p)
      byClinic.value.set(
        clinicId,
        result.items.map((p) => p.id),
      )
      return result
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

  function balanceByClinic(clinicId: string): number {
    return listByClinic(clinicId)
      .filter((p) => p.status === 'approved')
      .reduce((sum, p) => sum + p.amount, 0)
  }

  function pendingTotalByClinic(clinicId: string): number {
    return listByClinic(clinicId)
      .filter((p) => p.status === 'pending')
      .reduce((sum, p) => sum + p.amount, 0)
  }

  function receivablesByClinic(clinicId: string): { payment: Payment; settlementDate: string }[] {
    return listByClinic(clinicId)
      .filter((p) => p.status === 'pending')
      .map((p) => ({ payment: p, settlementDate: mockSettlementDate(p) }))
      .sort(
        (a, b) =>
          a.settlementDate.localeCompare(b.settlementDate) ||
          a.payment.id.localeCompare(b.payment.id),
      )
  }

  function upcomingReceivablesByClinic(clinicId: string, limit = 3) {
    return receivablesByClinic(clinicId).slice(0, limit)
  }

  return {
    items,
    byClinic,
    loading,
    error,
    listByClinic,
    create,
    fetchOne,
    list,
    evictByClinicId,
    balanceByClinic,
    pendingTotalByClinic,
    receivablesByClinic,
    upcomingReceivablesByClinic,
  }
})
