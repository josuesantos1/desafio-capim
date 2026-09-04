import { apiFetch } from './http'
import type { Payment, PaymentCreateInput } from '../types/api'

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
