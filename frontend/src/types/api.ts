export type ClinicStatus = 'pending' | 'active'

export interface Banking {
  bank: string
  agency: string
  account: string
}

export interface Address {
  street: string
  city: string
  state: string
  zip_code: string
}

export interface Clinic {
  id: string
  document: string
  legal_name: string
  trade_name: string
  banking: Banking | null
  description: string
  address: Address | null
  phone: string
  email: string
  website: string
  opening_hours: string
  specialties: string[]
  status: ClinicStatus
  created_at: string
  updated_at: string
}

export interface Dentist {
  id: string
  clinic_id: string
  name: string
  phone: string
  email: string
  bio: string
  specialties: string[]
  years_of_experience: number
  is_administrator: boolean
  is_legal_representative: boolean
  created_at: string
  updated_at: string
}

export type PaymentStatus = 'pending' | 'approved'

export interface Payment {
  id: string
  clinic_id: string
  dentist_id: string | null
  amount: number
  status: PaymentStatus
  pix_code: string
  created_at: string
  updated_at: string
  approved_at: string | null
}

export interface ProblemFieldError {
  field: string
  detail: string
}

export interface ProblemDetails {
  type: string
  title: string
  status: number
  detail?: string
  code?: string
  errors?: ProblemFieldError[]
}

export interface ClinicCreateInput {
  document: string
  legal_name: string
  trade_name: string
  email: string
  banking?: Banking
}

export interface ClinicUpdateInput {
  legal_name?: string
  trade_name?: string
  banking?: Banking
  document?: string
  description?: string
  address?: Address
  phone?: string
  email?: string
  website?: string
  opening_hours?: string
  specialties?: string[]
}

export interface DentistCreateInput {
  name: string
  phone: string
  email: string
  bio?: string
  specialties?: string[]
  years_of_experience?: number
}

export interface DentistUpdateInput {
  name?: string
  phone?: string
  email?: string
  bio?: string
  specialties?: string[]
  years_of_experience?: number
}

export interface DentistRolesInput {
  is_administrator?: boolean
  is_legal_representative?: boolean
}

export interface DentistListResult {
  items: Dentist[]
  total: number
  limit: number
  offset: number
}

export interface PaymentCreateInput {
  clinic_id: string
  amount: number
  dentist_id?: string
}

export interface PaymentListParams {
  limit?: number
  offset?: number
  status?: PaymentStatus
}

export interface PaymentListResult {
  items: Payment[]
  total: number
  limit: number
  offset: number
}

export interface ClinicListParams {
  limit?: number
  offset?: number
  q?: string
  city?: string
}

export interface ClinicListResult {
  items: Clinic[]
  total: number
  limit: number
  offset: number
}
