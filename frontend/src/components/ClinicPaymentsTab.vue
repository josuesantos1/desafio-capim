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

    <form v-if="clinic.status === 'active'" @submit.prevent="submit" class="form">
      <h2>Novo payment</h2>
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
      <button type="submit" :disabled="paymentsStore.loading" class="btn">Criar payment</button>
    </form>
    <p v-else>
      Esta clínica ainda não está ativa (precisa de ao menos um dentista administrador e um
      representante legal). Payments só podem ser criados com a clínica ativa.
    </p>

    <h2>Payments desta sessão</h2>
    <p v-if="paymentsStore.listByClinic(clinic.id).length === 0">
      Nenhum payment criado/consultado ainda nesta sessão.
    </p>
    <ul v-else class="list-plain">
      <li
        v-for="payment in paymentsStore.listByClinic(clinic.id)"
        :key="payment.id"
        class="list-item flex items-center gap-3"
      >
        <span>R$ {{ (payment.amount / 100).toFixed(2) }} — {{ payment.status }}</span>
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
