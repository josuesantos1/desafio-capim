import type { Clinic, Dentist } from '../types/api'

const SPECIALTIES = [
  'Ortodontia',
  'Implantodontia',
  'Odontopediatria',
  'Estética',
  'Endodontia',
  'Periodontia',
  'Prótese',
  'Clínica Geral',
]
const BIOS = [
  'Profissional dedicado, com foco em atendimento humanizado e resultados de longo prazo.',
  'Atua com atenção aos detalhes e atualização constante nas técnicas mais modernas da área.',
]

function hash(seed: string): number {
  let h = 0
  for (let i = 0; i < seed.length; i++) h += seed.charCodeAt(i)
  return h
}

function pick<T>(arr: T[], seed: string): T {
  return arr[hash(seed) % arr.length]
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

export interface DentistMockProfile {
  specialty: string
  bio: string
  specialties: string[]
  experienceYears: number
}

export function mockDentistProfile(dentist: Pick<Dentist, 'id'>): DentistMockProfile {
  const { id } = dentist
  const specialties = [pick(SPECIALTIES, id + ':spec1'), pick(SPECIALTIES, id + ':spec2')]

  return {
    specialty: specialties[0],
    bio: pick(BIOS, id + ':bio'),
    specialties,
    experienceYears: 2 + (hash(id + ':exp') % 20),
  }
}
