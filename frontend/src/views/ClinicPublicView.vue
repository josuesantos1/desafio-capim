<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useClinicsStore } from '../stores/clinics'
import { useDentistsStore } from '../stores/dentists'
import { mockClinicProfile } from '../mock/profile'

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
    <router-link to="/" class="text-sm text-neutral-500 hover:text-neutral-900 dark:hover:text-neutral-100">
      ← Início
    </router-link>

    <div v-if="clinicsStore.error" class="shell mt-6 max-w-md">
      <p class="card text-neutral-500 dark:text-neutral-400">Clínica não encontrada.</p>
    </div>

    <template v-else-if="clinic && mock">
      <div v-reveal class="shell mt-6">
        <div class="card overflow-hidden !p-0">
          <div
            class="flex h-36 items-end p-6"
            :style="{ background: mock.coverColor }"
          >
            <div
              class="flex h-16 w-16 items-center justify-center rounded-full bg-white text-2xl font-bold text-neutral-800 shadow-[0_8px_24px_-8px_rgba(0,0,0,0.35)]"
            >
              {{ mock.logoInitial }}
            </div>
          </div>
          <div class="flex flex-wrap items-center justify-between gap-4 p-6">
            <div>
              <h1 class="mt-0 mb-1 text-3xl">
                {{ clinic.trade_name }}
                <span v-if="clinic.status === 'pending'" class="badge">Em configuração</span>
              </h1>
              <p class="text-neutral-500 dark:text-neutral-400">
                <template v-if="clinic.address?.city">{{ clinic.address.city }} · </template>★
                {{ mock.rating }}
              </p>
            </div>
            <a v-if="clinic.email" :href="`mailto:${clinic.email}`">
              <button type="button" class="btn btn-primary group">
                Entrar em contato
                <span class="btn-icon !bg-white/20">↗</span>
              </button>
            </a>
            <button v-else type="button" class="btn" disabled>Contato não informado</button>
          </div>
        </div>
      </div>

      <div v-reveal class="mt-10 grid gap-6 md:grid-cols-2">
        <div class="shell">
          <div class="card">
            <h2 class="mt-0">Sobre</h2>
            <p class="text-neutral-500 dark:text-neutral-400">
              {{ clinic.description || 'Nenhuma descrição informada.' }}
            </p>

            <h2>Especialidades</h2>
            <div v-if="clinic.specialties.length" class="flex flex-wrap gap-2">
              <span v-for="s in clinic.specialties" :key="s" class="eyebrow normal-case">{{ s }}</span>
            </div>
            <p v-else class="text-neutral-500 dark:text-neutral-400">
              Nenhuma especialidade informada.
            </p>
          </div>
        </div>

        <div class="shell">
          <div class="card">
            <h2 class="mt-0">Informações</h2>
            <ul class="list-plain">
              <li class="list-item">
                Endereço:
                {{
                  clinic.address
                    ? `${clinic.address.street}, ${clinic.address.city} - ${clinic.address.state}`
                    : 'Endereço não informado'
                }}
              </li>
              <li class="list-item">Telefone: {{ clinic.phone || 'Não informado' }}</li>
              <li class="list-item">E-mail: {{ clinic.email || 'Não informado' }}</li>
              <li class="list-item">Website: {{ clinic.website || 'Não informado' }}</li>
              <li class="list-item">Horário: {{ clinic.opening_hours || 'Não informado' }}</li>
            </ul>
          </div>
        </div>
      </div>

      <div v-reveal class="mt-10">
        <h2 class="mt-0">Dentistas</h2>
        <p v-if="dentistsError" class="text-neutral-500 dark:text-neutral-400">
          Não foi possível carregar os dentistas.
        </p>
        <p v-else-if="dentistsStore.listByClinic(clinic.id).length === 0" class="text-neutral-500 dark:text-neutral-400">
          Nenhum dentista cadastrado.
        </p>
        <div v-else class="grid gap-4 sm:grid-cols-2">
          <router-link
            v-for="dentist in dentistsStore.listByClinic(clinic.id)"
            :key="dentist.id"
            :to="`/d/${clinic.id}/${dentist.id}`"
            class="shell no-underline"
          >
            <div class="card card-hover flex items-center justify-between">
              <div>
                <strong class="font-display">{{ dentist.name }}</strong>
                <p class="text-sm text-neutral-500 dark:text-neutral-400">
                  {{ dentist.specialties[0] ?? 'Clínico Geral' }}
                </p>
              </div>
              <span class="btn-icon !bg-neutral-900/5 dark:!bg-white/10">→</span>
            </div>
          </router-link>
        </div>
      </div>
    </template>
  </section>
</template>
