<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useClinicsStore } from '../stores/clinics'
import { useDentistsStore } from '../stores/dentists'
import { mockClinicProfile, mockDentistProfile } from '../mock/profile'
import type { Clinic, Dentist } from '../types/api'

const route = useRoute()
const clinicsStore = useClinicsStore()
const dentistsStore = useDentistsStore()

const term = computed(() => String(route.query.q ?? ''))
const loading = ref(false)

const clinicResults = ref<Clinic[]>([])
const dentistResults = ref<{ dentist: Dentist; clinic: Clinic }[]>([])

async function runSearch() {
  const q = term.value.trim().toLowerCase()
  loading.value = true

  await Promise.all(
    clinicsStore.list.map((clinic) =>
      dentistsStore.listByClinic(clinic.id).length === 0
        ? dentistsStore.list(clinic.id, 100).catch(() => {
            // seção de dentistas fica vazia para esta clínica; não bloqueia a busca
          })
        : Promise.resolve(),
    ),
  )

  const matchedClinics = clinicsStore.list.filter((clinic) => {
    const mock = mockClinicProfile(clinic)
    return (
      clinic.trade_name.toLowerCase().includes(q) ||
      mock.city.toLowerCase().includes(q) ||
      mock.specialties.some((s) => s.toLowerCase().includes(q))
    )
  })

  const matchedClinicIds = new Set(matchedClinics.map((c) => c.id))

  const matchedDentists: { dentist: Dentist; clinic: Clinic }[] = []
  for (const clinic of clinicsStore.list) {
    if (matchedClinicIds.has(clinic.id)) continue
    for (const dentist of dentistsStore.listByClinic(clinic.id)) {
      const mock = mockDentistProfile(dentist)
      if (dentist.name.toLowerCase().includes(q) || mock.specialty.toLowerCase().includes(q)) {
        matchedDentists.push({ dentist, clinic })
      }
    }
  }

  clinicResults.value = matchedClinics
  dentistResults.value = matchedDentists
  loading.value = false
}

watch(term, runSearch, { immediate: true })
</script>

<template>
  <section v-reveal>
    <span class="eyebrow">Resultados</span>
    <h1 class="mt-3">"{{ term }}"</h1>

    <div v-if="!loading && clinicResults.length === 0 && dentistResults.length === 0" class="shell max-w-md">
      <p class="card text-neutral-500 dark:text-neutral-400">Nenhum resultado encontrado.</p>
    </div>

    <div v-if="clinicResults.length" class="mb-10 grid gap-4 sm:grid-cols-2">
      <router-link
        v-for="clinic in clinicResults"
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
            ★ {{ mockClinicProfile(clinic).rating }} · {{ mockClinicProfile(clinic).city }}
          </p>
          <p class="text-sm text-neutral-500 dark:text-neutral-400">
            {{ mockClinicProfile(clinic).specialties.join(' · ') }}
          </p>
          <p class="text-xs tracking-wide text-neutral-400 uppercase dark:text-neutral-500">
            {{ dentistsStore.listByClinic(clinic.id).length }} dentista(s)
          </p>
        </div>
      </router-link>
    </div>

    <div v-if="dentistResults.length" class="grid gap-4 sm:grid-cols-2">
      <div v-for="{ dentist, clinic } in dentistResults" :key="dentist.id" class="shell">
        <div class="card card-hover flex flex-col gap-2">
          <strong class="font-display text-lg">{{ dentist.name }}</strong>
          <p class="text-sm text-neutral-500 dark:text-neutral-400">
            {{ mockDentistProfile(dentist).specialty }}
          </p>
          <div class="flex items-center gap-3 text-sm">
            <router-link :to="`/c/${clinic.id}`" class="text-neutral-500 hover:text-neutral-900 dark:hover:text-neutral-100">
              {{ clinic.trade_name }}
            </router-link>
            <router-link :to="`/d/${clinic.id}/${dentist.id}`" class="font-medium text-neutral-900 hover:underline dark:text-neutral-100">
              Ver perfil →
            </router-link>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
