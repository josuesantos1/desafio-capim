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
    <router-link :to="`/c/${clinicId}`" class="text-sm text-neutral-500 hover:text-neutral-900 dark:hover:text-neutral-100">
      ← Clínica
    </router-link>

    <div v-if="loadError" class="shell mt-6 max-w-md">
      <p class="card text-neutral-500 dark:text-neutral-400">Não foi possível carregar este perfil.</p>
    </div>
    <div v-else-if="!dentist" class="shell mt-6 max-w-md">
      <p class="card text-neutral-500 dark:text-neutral-400">Dentista não encontrado.</p>
    </div>

    <template v-else-if="mock">
      <div v-reveal class="mt-6">
        <span class="eyebrow">{{ mock.experienceYears }} anos de experiência</span>
        <h1 class="mt-3">{{ dentist.name }}</h1>
        <p class="text-lg text-neutral-500 dark:text-neutral-400">{{ mock.specialty }}</p>
      </div>

      <div v-reveal class="mt-8 grid gap-6 md:grid-cols-2">
        <div class="shell">
          <div class="card">
            <h2 class="mt-0">Sobre</h2>
            <p class="text-neutral-500 dark:text-neutral-400">{{ mock.bio }}</p>

            <h2>Especialidades</h2>
            <div class="flex flex-wrap gap-2">
              <span v-for="s in mock.specialties" :key="s" class="eyebrow normal-case">{{ s }}</span>
            </div>
          </div>
        </div>

        <div class="shell">
          <div class="card flex flex-col gap-4">
            <div>
              <h2 class="mt-0">Clínicas</h2>
              <router-link
                :to="`/c/${clinicId}`"
                class="font-medium text-neutral-900 hover:underline dark:text-neutral-100"
              >
                Ver clínica →
              </router-link>
            </div>
            <div>
              <h2>Contato</h2>
              <ul class="list-plain">
                <li class="list-item">Telefone: {{ dentist.phone }}</li>
                <li class="list-item">E-mail: {{ dentist.email }}</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </template>
  </section>
</template>
