<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useClinicsStore } from '../stores/clinics'
import { mockClinicProfile } from '../mock/profile'
import ErrorBanner from '../components/ErrorBanner.vue'

const router = useRouter()
const clinicsStore = useClinicsStore()

const query = ref('')
const location = ref('')

function submit() {
  router.push({ path: '/search', query: { q: query.value, city: location.value || undefined } })
}

onMounted(() => {
  clinicsStore.list({}).catch(() => {
    // erro já está em clinicsStore.error/resultsStatus
  })
})
</script>

<template>
  <section class="grid gap-16 md:grid-cols-2 md:items-start">
    <div v-reveal class="flex flex-col gap-6">
      <span class="eyebrow">Marketplace odontológico</span>
      <h1 class="text-5xl leading-[1.05] sm:text-6xl">
        Encontre a clínica ideal para o seu sorriso.
      </h1>
      <p class="max-w-md text-base text-neutral-500 dark:text-neutral-400">
        Descubra clínicas, conheça os dentistas e escolha com confiança — tudo em um só lugar.
      </p>

      <form @submit.prevent="submit" class="shell mt-2 max-w-md">
        <div class="card flex flex-col gap-3">
          <label class="field">
            Buscar clínica, dentista ou especialidade
            <input v-model="query" class="input" placeholder="Ex.: Ortodontia, Dra. Ana..." />
          </label>
          <label class="field">
            Localização
            <input v-model="location" placeholder="Cidade (opcional)" class="input" />
          </label>
          <button type="submit" class="btn btn-primary group self-start">
            Buscar
            <span class="btn-icon">↗</span>
          </button>
        </div>
      </form>
    </div>

    <div v-reveal class="flex flex-col gap-4">
      <h2 class="mt-0">Clínicas em destaque</h2>

      <div v-if="clinicsStore.resultsStatus === 'error'" class="shell">
        <ErrorBanner :problem="clinicsStore.error" />
      </div>

      <p v-else-if="clinicsStore.results.length === 0" class="shell">
        <span class="card block text-neutral-500 dark:text-neutral-400">
          Nenhuma clínica disponível ainda —
          <router-link
            to="/clinics"
            class="font-medium text-neutral-900 underline underline-offset-2 dark:text-neutral-100"
          >
            crie uma pela área de gestão
          </router-link>
          .
        </span>
      </p>

      <div v-else class="grid gap-4 sm:grid-cols-2">
        <router-link
          v-for="(clinic, i) in clinicsStore.results"
          :key="clinic.id"
          :to="`/c/${clinic.id}`"
          class="shell no-underline"
          :class="i === 0 ? 'sm:col-span-2' : ''"
        >
          <div class="card card-hover flex flex-col gap-1.5">
            <div class="flex items-center justify-between">
              <strong class="font-display text-lg text-neutral-900 dark:text-neutral-50">
                {{ clinic.trade_name }}
              </strong>
              <span v-if="clinic.status === 'pending'" class="badge">Em configuração</span>
            </div>
            <p class="text-sm text-neutral-500 dark:text-neutral-400">
              ★ {{ mockClinicProfile(clinic).rating }}
              <template v-if="clinic.address?.city"> · {{ clinic.address.city }}</template>
            </p>
          </div>
        </router-link>
      </div>
    </div>
  </section>
</template>
