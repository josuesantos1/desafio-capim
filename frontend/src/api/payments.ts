import { apiFetch } from './http'
import type { Payment, PaymentCreateInput, PaymentListParams, PaymentListResult } from '../types/api'

export function createPayment(
  input: PaymentCreateInput,
  idempotencyKey: string,
): Promise<Payment> {
  return apiFetch<Payment>('/payments', {
    method: 'POST',
    headers: { 'Idempotency-Key': idempotencyKey },
    body: JSON.stringify(input),
  })
}

export function getPayment(id: string): Promise<Payment> {
  return apiFetch<Payment>(`/payments/${id}`)
}

export function listPayments(
  clinicId: string,
  params: PaymentListParams = {},
): Promise<PaymentListResult> {
  const query = new URLSearchParams({ clinic_id: clinicId })
  if (params.limit) query.set('limit', String(params.limit))
  if (params.offset) query.set('offset', String(params.offset))
  if (params.status) query.set('status', params.status)
  return apiFetch<PaymentListResult>(`/payments?${query.toString()}`)
}
