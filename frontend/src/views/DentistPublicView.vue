<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useDentistsStore } from '../stores/dentists'
import { mockDentistProfile } from '../mock/profile'

const props = defineProps<{ clinicId: string; dentistId: string }>()

const dentistsStore = useDentistsStore()
const loadError = ref(false)

const dentist = computed(() => dentistsStore.items.get(props.dentistId) ?? null)
const mock = computed(() => (dentist.value ? mockDentistProfile(dentist.value) : null))

async function load() {
  loadError.value = false
  if (dentistsStore.listByClinic(props.clinicId).length === 0) {
    try {
      await dentistsStore.list(props.clinicId, 100)
    } catch {
      loadError.value = true
    }
  }
}

onMounted(load)
watch(() => [props.clinicId, props.dentistId], load)
</script>

<template>
  <section>
    <router-link :to="`/c/${clinicId}`" class="hover:underline">&larr; Clínica</router-link>

    <p v-if="loadError">Não foi possível carregar este perfil.</p>
    <p v-else-if="!dentist">Dentista não encontrado.</p>

    <template v-else-if="mock">
      <h1 class="mt-3">{{ dentist.name }}</h1>
      <p><strong>{{ mock.specialty }}</strong> — {{ mock.experienceYears }} anos de experiência</p>

      <h2>Sobre</h2>
      <p>{{ mock.bio }}</p>

      <h2>Especialidades</h2>
      <p>{{ mock.specialties.join(' · ') }}</p>

      <h2>Clínicas</h2>
      <router-link :to="`/c/${clinicId}`" class="hover:underline">Ver clínica</router-link>

      <h2>Contato</h2>
      <ul class="list-plain">
        <li>Telefone: {{ dentist.phone }}</li>
        <li>E-mail: {{ dentist.email }}</li>
      </ul>
    </template>
  </section>
</template>
