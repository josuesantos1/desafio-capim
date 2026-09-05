<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useClinicsStore } from '../stores/clinics'
import { mockClinicProfile } from '../mock/profile'
import ErrorBanner from './ErrorBanner.vue'
import type { Clinic, ClinicUpdateInput } from '../types/api'

const props = defineProps<{ clinic: Clinic }>()

const router = useRouter()
const store = useClinicsStore()

const form = reactive<Pick<ClinicUpdateInput, 'legal_name' | 'trade_name' | 'description' | 'phone' | 'email' | 'website' | 'opening_hours'>>({
  legal_name: props.clinic.legal_name,
  trade_name: props.clinic.trade_name,
  description: props.clinic.description,
  phone: props.clinic.phone,
  email: props.clinic.email,
  website: props.clinic.website,
  opening_hours: props.clinic.opening_hours,
})
const withBanking = ref(props.clinic.banking !== null)
const bank = ref(props.clinic.banking?.bank ?? '')
const agency = ref(props.clinic.banking?.agency ?? '')
const account = ref(props.clinic.banking?.account ?? '')

const street = ref(props.clinic.address?.street ?? '')
const city = ref(props.clinic.address?.city ?? '')
const state = ref(props.clinic.address?.state ?? '')
const zipCode = ref(props.clinic.address?.zip_code ?? '')
const specialtiesText = ref(props.clinic.specialties.join('\n'))

watch(
  () => props.clinic,
  (clinic) => {
    form.legal_name = clinic.legal_name
    form.trade_name = clinic.trade_name
    form.description = clinic.description
    form.phone = clinic.phone
    form.email = clinic.email
    form.website = clinic.website
    form.opening_hours = clinic.opening_hours
    withBanking.value = clinic.banking !== null
    bank.value = clinic.banking?.bank ?? ''
    agency.value = clinic.banking?.agency ?? ''
    account.value = clinic.banking?.account ?? ''
    street.value = clinic.address?.street ?? ''
    city.value = clinic.address?.city ?? ''
    state.value = clinic.address?.state ?? ''
    zipCode.value = clinic.address?.zip_code ?? ''
    specialtiesText.value = clinic.specialties.join('\n')
  },
)

async function submit() {
  const specialties = [
    ...new Set(
      specialtiesText.value
        .split('\n')
        .map((s) => s.trim())
        .filter((s) => s !== ''),
    ),
  ]

  const input: ClinicUpdateInput = {
    legal_name: form.legal_name,
    trade_name: form.trade_name,
    description: form.description,
    phone: form.phone,
    email: form.email,
    website: form.website,
    opening_hours: form.opening_hours,
    address: { street: street.value, city: city.value, state: state.value, zip_code: zipCode.value },
    specialties,
    ...(withBanking.value
      ? { banking: { bank: bank.value, agency: agency.value, account: account.value } }
      : {}),
  }
  try {
    await store.update(props.clinic.id, input)
  } catch {
    // erro já está em store.error
  }
}

async function remove() {
  if (!confirm('Excluir esta clínica? Esta ação não pode ser desfeita.')) return
  try {
    await store.remove(props.clinic.id)
    router.push('/clinics')
  } catch {
    // erro já está em store.error
  }
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="shell max-w-md">
      <form @submit.prevent="submit" class="card !mb-0 flex flex-col gap-3">
        <ErrorBanner :problem="store.error" />

        <div class="mb-1 flex items-center gap-3">
          <div
            class="flex h-12 w-12 items-center justify-center rounded-full bg-neutral-900/5 text-lg font-bold text-neutral-700 dark:bg-white/10 dark:text-neutral-200"
          >
            {{ mockClinicProfile(clinic).logoInitial }}
          </div>
          <p class="text-sm text-neutral-500 dark:text-neutral-400">Status: {{ clinic.status }}</p>
        </div>

        <label class="field">
          Documento
          <input :value="clinic.document" disabled class="input" />
        </label>
        <label class="field">
          Razão social
          <input v-model="form.legal_name" required class="input" />
        </label>
        <label class="field">
          Nome fantasia
          <input v-model="form.trade_name" required class="input" />
        </label>
        <label class="field">
          Descrição
          <textarea v-model="form.description" rows="3" class="input" />
        </label>
        <label class="field">
          Telefone
          <input v-model="form.phone" class="input" />
        </label>
        <label class="field">
          E-mail
          <input v-model="form.email" type="email" required class="input" />
        </label>
        <label class="field">
          Website
          <input v-model="form.website" class="input" />
        </label>
        <label class="field">
          Horário de funcionamento
          <input v-model="form.opening_hours" class="input" placeholder="Ex.: Seg a Sex, 8h às 18h" />
        </label>

        <fieldset class="flex flex-col gap-3 border-0 p-0">
          <legend class="text-sm font-medium text-neutral-700 dark:text-neutral-300">Endereço</legend>
          <label class="field">
            Rua
            <input v-model="street" class="input" />
          </label>
          <label class="field">
            Cidade
            <input v-model="city" class="input" />
          </label>
          <label class="field">
            Estado
            <input v-model="state" class="input" />
          </label>
          <label class="field">
            CEP
            <input v-model="zipCode" class="input" />
          </label>
        </fieldset>

        <label class="field">
          Especialidades (uma por linha)
          <textarea v-model="specialtiesText" rows="3" class="input" />
        </label>

        <label class="field-inline">
          <input v-model="withBanking" type="checkbox" />
          Dados bancários
        </label>
        <fieldset v-if="withBanking" class="flex flex-col gap-3 border-0 p-0">
          <label class="field">
            Banco
            <input v-model="bank" required class="input" />
          </label>
          <label class="field">
            Agência
            <input v-model="agency" required class="input" />
          </label>
          <label class="field">
            Conta
            <input v-model="account" required class="input" />
          </label>
        </fieldset>

        <button type="submit" :disabled="store.loading" class="btn btn-primary self-start">
          Salvar
        </button>
      </form>
    </div>

    <button type="button" :disabled="store.loading" class="btn self-start" @click="remove">
      Excluir clínica
    </button>
  </div>
</template>
