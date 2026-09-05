<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useClinicsStore } from '../stores/clinics'
import ErrorBanner from '../components/ErrorBanner.vue'
import type { ClinicCreateInput } from '../types/api'

const router = useRouter()
const store = useClinicsStore()

const form = reactive<ClinicCreateInput>({
  document: '',
  legal_name: '',
  trade_name: '',
})
const withBanking = ref(false)
const bank = ref('')
const agency = ref('')
const account = ref('')

async function submit() {
  const input: ClinicCreateInput = {
    document: form.document,
    legal_name: form.legal_name,
    trade_name: form.trade_name,
    ...(withBanking.value
      ? { banking: { bank: bank.value, agency: agency.value, account: account.value } }
      : {}),
  }
  try {
    const clinic = await store.create(input)
    router.push(`/clinics/${clinic.id}`)
  } catch {
    // erro já está em store.error, exibido pelo ErrorBanner
  }
}
</script>

<template>
  <section>
    <h1>Clínicas</h1>

    <form @submit.prevent="submit" class="form">
      <h2>Nova clínica</h2>
      <ErrorBanner :problem="store.error" />
      <label class="field">
        Documento (CPF/CNPJ)
        <input v-model="form.document" required class="input" />
      </label>
      <label class="field">
        Razão social
        <input v-model="form.legal_name" required class="input" />
      </label>
      <label class="field">
        Nome fantasia
        <input v-model="form.trade_name" required class="input" />
      </label>
      <label class="field-inline">
        <input v-model="withBanking" type="checkbox" />
        Informar dados bancários
      </label>
      <fieldset v-if="withBanking" class="flex flex-col gap-2 border-0 p-0">
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
      <button type="submit" :disabled="store.loading" class="btn">Criar clínica</button>
    </form>

    <h2>Clínicas desta sessão</h2>
    <p v-if="store.list.length === 0">Nenhuma clínica criada/consultada ainda nesta sessão.</p>
    <ul v-else class="list-plain">
      <li v-for="clinic in store.list" :key="clinic.id" class="list-item">
        <router-link :to="`/clinics/${clinic.id}`" class="hover:underline">
          {{ clinic.trade_name }} — {{ clinic.status }}
        </router-link>
      </li>
    </ul>
  </section>
</template>
