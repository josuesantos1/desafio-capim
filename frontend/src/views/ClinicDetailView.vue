<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useClinicsStore } from '../stores/clinics'
import ErrorBanner from '../components/ErrorBanner.vue'
import ClinicDadosTab from '../components/ClinicDadosTab.vue'
import ClinicDentistasTab from '../components/ClinicDentistasTab.vue'
import ClinicPaymentsTab from '../components/ClinicPaymentsTab.vue'

const props = defineProps<{ id: string }>()

const store = useClinicsStore()
const activeTab = ref<'dados' | 'dentistas' | 'payments'>('dados')

const clinic = computed(() => store.items.get(props.id) ?? null)

async function load() {
  activeTab.value = 'dados'
  try {
    await store.fetch(props.id)
  } catch {
    // erro já está em store.error
  }
}

onMounted(load)
watch(() => props.id, load)
</script>

<template>
  <section v-reveal>
    <router-link to="/clinics" class="text-sm text-neutral-500 hover:text-neutral-900 dark:hover:text-neutral-100">
      ← Clínicas
    </router-link>

    <div v-if="clinic" class="mt-4 flex flex-wrap items-center justify-between gap-3">
      <h1 class="mb-0">{{ clinic.trade_name }}</h1>
      <router-link
        :to="`/c/${clinic.id}`"
        class="text-sm font-medium text-neutral-500 hover:text-neutral-900 dark:hover:text-neutral-100"
      >
        Ver perfil público →
      </router-link>
    </div>
    <ErrorBanner :problem="store.error" />

    <nav v-if="clinic" class="my-6 inline-flex w-max gap-1 rounded-full border border-neutral-900/10 bg-white p-1 dark:border-white/10 dark:bg-neutral-900">
      <button
        :class="[
          'rounded-full px-4 py-1.5 text-sm transition-all duration-300 ease-[cubic-bezier(0.32,0.72,0,1)]',
          activeTab === 'dados'
            ? 'bg-neutral-900 text-white dark:bg-white dark:text-neutral-900'
            : 'text-neutral-500 hover:text-neutral-900 dark:hover:text-neutral-100',
        ]"
        @click="activeTab = 'dados'"
      >
        Dados
      </button>
      <button
        :disabled="store.loading"
        :class="[
          'rounded-full px-4 py-1.5 text-sm transition-all duration-300 ease-[cubic-bezier(0.32,0.72,0,1)] disabled:cursor-not-allowed disabled:opacity-50',
          activeTab === 'dentistas'
            ? 'bg-neutral-900 text-white dark:bg-white dark:text-neutral-900'
            : 'text-neutral-500 hover:text-neutral-900 dark:hover:text-neutral-100',
        ]"
        @click="activeTab = 'dentistas'"
      >
        Dentistas
      </button>
      <button
        :disabled="store.loading"
        :class="[
          'rounded-full px-4 py-1.5 text-sm transition-all duration-300 ease-[cubic-bezier(0.32,0.72,0,1)] disabled:cursor-not-allowed disabled:opacity-50',
          activeTab === 'payments'
            ? 'bg-neutral-900 text-white dark:bg-white dark:text-neutral-900'
            : 'text-neutral-500 hover:text-neutral-900 dark:hover:text-neutral-100',
        ]"
        @click="activeTab = 'payments'"
      >
        Payments
      </button>
    </nav>

    <ClinicDadosTab v-if="clinic && activeTab === 'dados'" :clinic="clinic" />
    <ClinicDentistasTab v-if="clinic && activeTab === 'dentistas'" :clinic-id="clinic.id" />
    <ClinicPaymentsTab v-if="clinic && activeTab === 'payments'" :clinic="clinic" />
  </section>
</template>
