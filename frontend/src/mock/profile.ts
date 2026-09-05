import type { Clinic, Dentist } from '../types/api'

const CITIES = [
  'São Paulo, SP',
  'Rio de Janeiro, RJ',
  'Belo Horizonte, MG',
  'Curitiba, PR',
  'Porto Alegre, RS',
  'Salvador, BA',
]
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
const DESCRIPTIONS = [
  'Clínica odontológica moderna, com foco em conforto e atendimento humanizado.',
  'Equipe experiente dedicada a tratamentos de excelência e resultados duradouros.',
  'Estrutura completa para toda a família, do check-up de rotina a procedimentos especializados.',
]
const BIOS = [
  'Profissional dedicado, com foco em atendimento humanizado e resultados de longo prazo.',
  'Atua com atenção aos detalhes e atualização constante nas técnicas mais modernas da área.',
]
const HOURS = ['Seg a Sex, 8h às 18h', 'Seg a Sáb, 9h às 19h', 'Seg a Sex, 7h às 17h']

function hash(seed: string): number {
  let h = 0
  for (let i = 0; i < seed.length; i++) h += seed.charCodeAt(i)
  return h
}

function pick<T>(arr: T[], seed: string): T {
  return arr[hash(seed) % arr.length]
}

export interface ClinicMockProfile {
  city: string
  rating: number
  specialties: string[]
  description: string
  address: string
  phone: string
  email: string
  website: string
  hours: string
  coverColor: string
  logoInitial: string
}

export function mockClinicProfile(clinic: Pick<Clinic, 'id' | 'trade_name'>): ClinicMockProfile {
  const { id, trade_name } = clinic
  const rating = 3.5 + (hash(id + ':rating') % 16) / 10
  const specialties = [pick(SPECIALTIES, id + ':spec1'), pick(SPECIALTIES, id + ':spec2')]
  const hue = hash(id + ':color') % 360
  const slug = trade_name.toLowerCase().replace(/[^a-z0-9]+/g, '')

  return {
    city: pick(CITIES, id + ':city'),
    rating: Math.round(rating * 10) / 10,
    specialties,
    description: pick(DESCRIPTIONS, id + ':desc'),
    address: `Av. Exemplo, ${100 + (hash(id + ':addr') % 900)}`,
    phone: `(11) ${4000 + (hash(id + ':phone') % 6000)}-0000`,
    email: `contato@${slug || 'clinica'}.com.br`,
    website: `www.${slug || 'clinica'}.com.br`,
    hours: pick(HOURS, id + ':hours'),
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
