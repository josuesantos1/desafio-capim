<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useClinicsStore } from '../stores/clinics'
import { useDentistsStore } from '../stores/dentists'
import { mockClinicProfile, mockDentistProfile } from '../mock/profile'

const props = defineProps<{ id: string }>()

const clinicsStore = useClinicsStore()
const dentistsStore = useDentistsStore()
const dentistsError = ref(false)

const clinic = computed(() => clinicsStore.items.get(props.id) ?? null)
const mock = computed(() => (clinic.value ? mockClinicProfile(clinic.value) : null))

async function load() {
  dentistsError.value = false
  try {
    await clinicsStore.fetch(props.id)
  } catch {
    return
  }
  try {
    await dentistsStore.list(props.id, 100)
  } catch {
    dentistsError.value = true
  }
}

onMounted(load)
watch(() => props.id, load)
</script>

<template>
  <section>
    <router-link to="/" class="hover:underline">&larr; Início</router-link>

    <p v-if="clinicsStore.error">Clínica não encontrada.</p>

    <template v-else-if="clinic && mock">
      <div
        class="mt-4 flex h-30 items-end rounded-lg p-2"
        :style="{ background: mock.coverColor }"
      >
        <div
          class="flex h-12 w-12 items-center justify-center rounded-full bg-white text-xl font-bold text-slate-700"
        >
          {{ mock.logoInitial }}
        </div>
      </div>
      <h1 class="mt-3">
        {{ clinic.trade_name }}
        <span v-if="clinic.status === 'pending'" class="badge">Em configuração</span>
      </h1>
      <p>{{ mock.city }} — ★ {{ mock.rating }}</p>
      <a :href="`mailto:${mock.email}`">
        <button type="button" class="btn">Entrar em contato</button>
      </a>

      <h2>Sobre</h2>
      <p>{{ mock.description }}</p>

      <h2>Informações</h2>
      <ul class="list-plain">
        <li>Endereço: {{ mock.address }}</li>
        <li>Telefone: {{ mock.phone }}</li>
        <li>E-mail: {{ mock.email }}</li>
        <li>Website: {{ mock.website }}</li>
        <li>Horário: {{ mock.hours }}</li>
      </ul>

      <h2>Especialidades</h2>
      <p>{{ mock.specialties.join(' · ') }}</p>

      <h2>Dentistas</h2>
      <p v-if="dentistsError">Não foi possível carregar os dentistas.</p>
      <p v-else-if="dentistsStore.listByClinic(clinic.id).length === 0">
        Nenhum dentista cadastrado.
      </p>
      <ul v-else class="list-plain">
        <li v-for="dentist in dentistsStore.listByClinic(clinic.id)" :key="dentist.id" class="list-item">
          {{ dentist.name }} — {{ mockDentistProfile(dentist).specialty }}
          <router-link :to="`/d/${clinic.id}/${dentist.id}`" class="hover:underline">
            Ver perfil
          </router-link>
        </li>
      </ul>
    </template>
  </section>
</template>
