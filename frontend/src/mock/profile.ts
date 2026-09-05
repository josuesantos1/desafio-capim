import type { Clinic } from '../types/api'

export function hash(seed: string): number {
  let h = 0
  for (let i = 0; i < seed.length; i++) h += seed.charCodeAt(i)
  return h
}

export interface ClinicMockProfile {
  rating: number
  coverColor: string
  logoInitial: string
}

export function mockClinicProfile(clinic: Pick<Clinic, 'id' | 'trade_name'>): ClinicMockProfile {
  const { id, trade_name } = clinic
  const rating = 3.5 + (hash(id + ':rating') % 16) / 10
  const hue = hash(id + ':color') % 360

  return {
    rating: Math.round(rating * 10) / 10,
    coverColor: `hsl(${hue}, 60%, 55%)`,
    logoInitial: trade_name.charAt(0).toUpperCase(),
  }
}
