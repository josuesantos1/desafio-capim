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
  <section>
    <h1>Resultados para "{{ term }}"</h1>

    <p v-if="!loading && clinicResults.length === 0 && dentistResults.length === 0">
      Nenhum resultado encontrado.
    </p>

    <ul v-if="clinicResults.length" class="list-plain mb-6">
      <li v-for="clinic in clinicResults" :key="clinic.id" class="list-item">
        <router-link :to="`/c/${clinic.id}`" class="hover:underline">
          <strong>{{ clinic.trade_name }}</strong>
          <span v-if="clinic.status === 'pending'" class="badge">Em configuração</span>
        </router-link>
        <br />
        ★ {{ mockClinicProfile(clinic).rating }} — {{ mockClinicProfile(clinic).city }}
        <br />
        Especialidades: {{ mockClinicProfile(clinic).specialties.join(' · ') }}
        <br />
        {{ dentistsStore.listByClinic(clinic.id).length }} dentista(s)
      </li>
    </ul>

    <ul v-if="dentistResults.length" class="list-plain">
      <li v-for="{ dentist, clinic } in dentistResults" :key="dentist.id" class="list-item">
        <strong>{{ dentist.name }}</strong> — {{ mockDentistProfile(dentist).specialty }}
        <br />
        <router-link :to="`/c/${clinic.id}`" class="hover:underline">{{
          clinic.trade_name
        }}</router-link>
        —
        <router-link :to="`/d/${clinic.id}/${dentist.id}`" class="hover:underline">
          Ver perfil
        </router-link>
      </li>
    </ul>
  </section>
</template>
