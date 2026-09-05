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
  <section>
    <router-link to="/clinics" class="hover:underline">&larr; Clínicas</router-link>

    <h1 v-if="clinic">
      {{ clinic.trade_name }}
      <router-link :to="`/c/${clinic.id}`" class="ml-3 text-sm font-normal hover:underline">
        Ver perfil público
      </router-link>
    </h1>
    <ErrorBanner :problem="store.error" />

    <nav v-if="clinic" class="my-4 flex gap-2">
      <button
        :class="[
          'cursor-pointer border-0 bg-transparent p-0',
          activeTab === 'dados' ? 'font-bold underline' : '',
        ]"
        @click="activeTab = 'dados'"
      >
        Dados
      </button>
      <button
        :disabled="store.loading"
        :class="[
          'cursor-pointer border-0 bg-transparent p-0 disabled:cursor-not-allowed disabled:opacity-50',
          activeTab === 'dentistas' ? 'font-bold underline' : '',
        ]"
        @click="activeTab = 'dentistas'"
      >
        Dentistas
      </button>
      <button
        :disabled="store.loading"
        :class="[
          'cursor-pointer border-0 bg-transparent p-0 disabled:cursor-not-allowed disabled:opacity-50',
          activeTab === 'payments' ? 'font-bold underline' : '',
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
