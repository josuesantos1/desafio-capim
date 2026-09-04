<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useClinicsStore } from '../stores/clinics'
import ErrorBanner from './ErrorBanner.vue'
import type { Clinic, ClinicUpdateInput } from '../types/api'

const props = defineProps<{ clinic: Clinic }>()

const router = useRouter()
const store = useClinicsStore()

const form = reactive<ClinicUpdateInput>({
  legal_name: props.clinic.legal_name,
  trade_name: props.clinic.trade_name,
})
const withBanking = ref(props.clinic.banking !== null)
const bank = ref(props.clinic.banking?.bank ?? '')
const agency = ref(props.clinic.banking?.agency ?? '')
const account = ref(props.clinic.banking?.account ?? '')

watch(
  () => props.clinic,
  (clinic) => {
    form.legal_name = clinic.legal_name
    form.trade_name = clinic.trade_name
    withBanking.value = clinic.banking !== null
    bank.value = clinic.banking?.bank ?? ''
    agency.value = clinic.banking?.agency ?? ''
    account.value = clinic.banking?.account ?? ''
  },
)

async function submit() {
  const input: ClinicUpdateInput = {
    legal_name: form.legal_name,
    trade_name: form.trade_name,
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
  <div>
    <form @submit.prevent="submit">
      <ErrorBanner :problem="store.error" />
      <label>
        Documento
        <input :value="clinic.document" disabled />
      </label>
      <label>
        Razão social
        <input v-model="form.legal_name" required />
      </label>
      <label>
        Nome fantasia
        <input v-model="form.trade_name" required />
      </label>
      <label>
        <input v-model="withBanking" type="checkbox" />
        Dados bancários
      </label>
      <fieldset v-if="withBanking">
        <label>
          Banco
          <input v-model="bank" required />
        </label>
        <label>
          Agência
          <input v-model="agency" required />
        </label>
        <label>
          Conta
          <input v-model="account" required />
        </label>
      </fieldset>
      <p>Status: {{ clinic.status }}</p>
      <button type="submit" :disabled="store.loading">Salvar</button>
    </form>

    <button type="button" :disabled="store.loading" @click="remove">Excluir clínica</button>
  </div>
</template>
