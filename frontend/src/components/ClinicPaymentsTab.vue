<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useDentistsStore } from '../stores/dentists'
import { usePaymentsStore } from '../stores/payments'
import ErrorBanner from './ErrorBanner.vue'
import type { Clinic } from '../types/api'

const props = defineProps<{ clinic: Clinic }>()

const dentistsStore = useDentistsStore()
const paymentsStore = usePaymentsStore()

const amountReais = ref('')
const dentistId = ref('')

onMounted(() => {
  if (dentistsStore.listByClinic(props.clinic.id).length === 0) {
    dentistsStore.list(props.clinic.id).catch(() => {
      // dropdown de dentista fica vazio; não é uma falha crítica desta aba
    })
  }
})

function formatCurrency(cents: number): string {
  return (cents / 100).toFixed(2)
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('pt-BR')
}

async function submit() {
  const amount = Math.round(Number(amountReais.value.replace(',', '.')) * 100)
  try {
    await paymentsStore.create({
      clinic_id: props.clinic.id,
      amount,
      ...(dentistId.value ? { dentist_id: dentistId.value } : {}),
    })
    amountReais.value = ''
    dentistId.value = ''
  } catch {
    // erro já está em paymentsStore.error
  }
}

function refresh(paymentId: string) {
  paymentsStore.fetchOne(paymentId).catch(() => {
    // erro já está em paymentsStore.error
  })
}
</script>

<template>
  <div>
    <ErrorBanner :problem="paymentsStore.error" />

    <div class="shell mb-6">
      <div class="card grid gap-4 sm:grid-cols-2">
        <div>
          <p class="text-sm text-neutral-500 dark:text-neutral-400">Saldo disponível</p>
          <p class="font-display text-2xl">
            R$ {{ formatCurrency(paymentsStore.balanceByClinic(clinic.id)) }}
          </p>
        </div>
        <div>
          <p class="text-sm text-neutral-500 dark:text-neutral-400">A receber</p>
          <p class="font-display text-2xl">
            R$ {{ formatCurrency(paymentsStore.pendingTotalByClinic(clinic.id)) }}
          </p>
        </div>
        <p class="text-xs text-neutral-400 sm:col-span-2 dark:text-neutral-500">
          Baseado nos payments consultados nesta sessão.
        </p>
      </div>
    </div>

    <div v-if="clinic.status === 'active'" class="shell max-w-md">
      <form @submit.prevent="submit" class="card !mb-0 flex flex-col gap-3">
        <h2 class="mt-0">Novo payment</h2>
        <label class="field">
          Valor (R$)
          <input
            v-model="amountReais"
            required
            inputmode="decimal"
            placeholder="0,00"
            class="input"
          />
        </label>
        <label class="field">
          Dentista (opcional)
          <select v-model="dentistId" class="input">
            <option value="">— nenhum —</option>
            <option
              v-for="dentist in dentistsStore.listByClinic(clinic.id)"
              :key="dentist.id"
              :value="dentist.id"
            >
              {{ dentist.name }}
            </option>
          </select>
        </label>
        <button type="submit" :disabled="paymentsStore.loading" class="btn btn-primary self-start">
          Criar payment
        </button>
      </form>
    </div>
    <p v-else class="max-w-md text-neutral-500 dark:text-neutral-400">
      Esta clínica ainda não está ativa (precisa de ao menos um dentista administrador e um
      representante legal). Payments só podem ser criados com a clínica ativa.
    </p>

    <h2>Próximos recebimentos</h2>
    <p
      v-if="paymentsStore.upcomingReceivablesByClinic(clinic.id).length === 0"
      class="text-neutral-500 dark:text-neutral-400"
    >
      Nenhum recebível no momento.
    </p>
    <ul v-else class="list-plain">
      <li
        v-for="{ payment, settlementDate } in paymentsStore.upcomingReceivablesByClinic(clinic.id)"
        :key="payment.id"
        class="list-item"
      >
        R$ {{ formatCurrency(payment.amount) }} — {{ formatDate(settlementDate) }} (estimado)
      </li>
    </ul>

    <h2>Recebíveis</h2>
    <p
      v-if="paymentsStore.receivablesByClinic(clinic.id).length === 0"
      class="text-neutral-500 dark:text-neutral-400"
    >
      Nenhum recebível no momento.
    </p>
    <ul v-else class="list-plain">
      <li
        v-for="{ payment, settlementDate } in paymentsStore.receivablesByClinic(clinic.id)"
        :key="payment.id"
        class="list-item"
      >
        R$ {{ formatCurrency(payment.amount) }} — {{ formatDate(settlementDate) }} (estimado) — Pendente
      </li>
    </ul>

    <h2>Histórico de payments</h2>
    <p
      v-if="paymentsStore.listByClinic(clinic.id).length === 0"
      class="text-neutral-500 dark:text-neutral-400"
    >
      Nenhum payment criado/consultado ainda nesta sessão.
    </p>
    <ul v-else class="list-plain">
      <li
        v-for="payment in paymentsStore.listByClinic(clinic.id)"
        :key="payment.id"
        class="list-item flex items-center gap-3"
      >
        <span>R$ {{ formatCurrency(payment.amount) }} — {{ payment.status }}</span>
        <button
          type="button"
          :disabled="paymentsStore.loading"
          class="btn"
          @click="refresh(payment.id)"
        >
          Atualizar status
        </button>
      </li>
    </ul>
  </div>
</template>
