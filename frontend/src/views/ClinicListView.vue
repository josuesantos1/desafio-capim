<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useClinicsStore } from '../stores/clinics'
import { useAuthStore } from '../stores/auth'
import ErrorBanner from '../components/ErrorBanner.vue'
import type { ClinicCreateInput } from '../types/api'

const router = useRouter()
const store = useClinicsStore()
const authStore = useAuthStore()

const form = reactive<ClinicCreateInput>({
  document: '',
  legal_name: '',
  trade_name: '',
  email: '',
})
const withBanking = ref(false)
const bank = ref('')
const agency = ref('')
const account = ref('')

const myClinics = computed(() => store.results.filter((c) => store.isOwner(c, authStore.email)))

onMounted(() => {
  store.list({ limit: 100 }).catch(() => {
    // erro já está em store.error/resultsStatus
  })
})

async function submit() {
  const input: ClinicCreateInput = {
    document: form.document,
    legal_name: form.legal_name,
    trade_name: form.trade_name,
    email: form.email,
    ...(withBanking.value
      ? { banking: { bank: bank.value, agency: agency.value, account: account.value } }
      : {}),
  }
  try {
    const clinic = await store.create(input)
    store.list({ limit: 100 }).catch(() => {
      // erro já está em store.error/resultsStatus
    })
    router.push(`/clinics/${clinic.id}`)
  } catch {
    // erro já está em store.error, exibido pelo ErrorBanner
  }
}
</script>

<template>
  <section v-reveal>
    <span class="eyebrow">Área de gestão</span>
    <h1 class="mt-3">Clínicas</h1>

    <div class="shell max-w-md">
      <form @submit.prevent="submit" class="card !mb-0 flex flex-col gap-3">
        <h2 class="mt-0">Nova clínica</h2>
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
        <label class="field">
          E-mail
          <input v-model="form.email" type="email" required class="input" />
        </label>
        <label class="field-inline">
          <input v-model="withBanking" type="checkbox" />
          Informar dados bancários
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
          Criar clínica
        </button>
      </form>
    </div>

    <h2>Minhas clínicas</h2>
    <div v-if="store.resultsStatus === 'error'" class="shell max-w-md">
      <ErrorBanner :problem="store.error" />
    </div>
    <p v-else-if="myClinics.length === 0" class="text-neutral-500 dark:text-neutral-400">
      Nenhuma clínica sua cadastrada ainda.
    </p>
    <div v-else class="grid gap-4 sm:grid-cols-2">
      <router-link
        v-for="clinic in myClinics"
        :key="clinic.id"
        :to="`/clinics/${clinic.id}`"
        class="shell no-underline"
      >
        <div class="card card-hover flex items-center justify-between">
          <span class="font-medium text-neutral-900 dark:text-neutral-100">
            {{ clinic.trade_name }}
          </span>
          <span class="text-sm text-neutral-500 dark:text-neutral-400">{{ clinic.status }}</span>
        </div>
      </router-link>
    </div>
  </section>
</template>
