export type ClinicStatus = 'pending' | 'active'

export interface Banking {
  bank: string
  agency: string
  account: string
}

export interface Clinic {
  id: string
  document: string
  legal_name: string
  trade_name: string
  banking: Banking | null
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
  banking?: Banking
}

export interface ClinicUpdateInput {
  legal_name?: string
  trade_name?: string
  banking?: Banking
}

export interface DentistCreateInput {
  name: string
  phone: string
  email: string
}

export interface DentistUpdateInput {
  name?: string
  phone?: string
  email?: string
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
