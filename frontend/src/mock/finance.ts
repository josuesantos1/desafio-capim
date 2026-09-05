import type { Payment } from '../types/api'
import { hash } from './profile'

const SETTLEMENT_OFFSETS_DAYS = [7, 14, 21]

export function mockSettlementDate(payment: Pick<Payment, 'id' | 'created_at'>): string {
  const offsetDays =
    SETTLEMENT_OFFSETS_DAYS[hash(payment.id + ':settlement') % SETTLEMENT_OFFSETS_DAYS.length]
  const date = new Date(payment.created_at)
  date.setDate(date.getDate() + offsetDays)
  return date.toISOString()
}
