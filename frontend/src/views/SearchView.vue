<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useClinicsStore } from '../stores/clinics'
import { useDentistsStore } from '../stores/dentists'
import { mockClinicProfile } from '../mock/profile'
import ErrorBanner from '../components/ErrorBanner.vue'

const route = useRoute()
const clinicsStore = useClinicsStore()
const dentistsStore = useDentistsStore()

const term = computed(() => String(route.query.q ?? ''))
const city = computed(() => String(route.query.city ?? ''))

async function runSearch() {
  try {
    await clinicsStore.list({ q: term.value || undefined, city: city.value || undefined })
  } catch {
    return // erro já está em clinicsStore.error/resultsStatus
  }

  await Promise.all(
    clinicsStore.results.map((clinic) =>
      dentistsStore.listByClinic(clinic.id).length === 0
        ? dentistsStore.list(clinic.id, 100).catch(() => {
            // contagem de dentistas fica indisponível para esta clínica; não bloqueia a busca
          })
        : Promise.resolve(),
    ),
  )
}

watch([term, city], runSearch, { immediate: true })
</script>

<template>
  <section v-reveal>
    <span class="eyebrow">Resultados</span>
    <h1 class="mt-3">"{{ term }}"</h1>

    <div v-if="clinicsStore.resultsStatus === 'error'" class="shell max-w-md">
      <ErrorBanner :problem="clinicsStore.error" />
    </div>

    <div v-else-if="clinicsStore.results.length === 0" class="shell max-w-md">
      <p class="card text-neutral-500 dark:text-neutral-400">Nenhum resultado encontrado.</p>
    </div>

    <div v-else class="grid gap-4 sm:grid-cols-2">
      <router-link
        v-for="clinic in clinicsStore.results"
        :key="clinic.id"
        :to="`/c/${clinic.id}`"
        class="shell no-underline"
      >
        <div class="card card-hover flex flex-col gap-2">
          <div class="flex items-center justify-between">
            <strong class="font-display text-lg">{{ clinic.trade_name }}</strong>
            <span v-if="clinic.status === 'pending'" class="badge">Em configuração</span>
          </div>
          <p class="text-sm text-neutral-500 dark:text-neutral-400">
            ★ {{ mockClinicProfile(clinic).rating }}
            <template v-if="clinic.address?.city"> · {{ clinic.address.city }}</template>
          </p>
          <p v-if="clinic.specialties.length" class="text-sm text-neutral-500 dark:text-neutral-400">
            {{ clinic.specialties.join(' · ') }}
          </p>
          <p class="text-xs tracking-wide text-neutral-400 uppercase dark:text-neutral-500">
            {{ dentistsStore.listByClinic(clinic.id).length }} dentista(s)
          </p>
        </div>
      </router-link>
    </div>
  </section>
</template>
