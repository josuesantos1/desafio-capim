<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const step = ref<'email' | 'code'>('email')
const email = ref('')
const code = ref('')
const codeError = ref('')

function sendCode() {
  step.value = 'code'
  codeError.value = ''
}

function submitCode() {
  if (code.value.trim() === '00000') {
    authStore.login(email.value)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/clinics'
    router.push(redirect)
  } else {
    codeError.value = 'Código inválido.'
  }
}
</script>

<template>
  <section class="mx-auto max-w-sm">
    <span class="eyebrow">Entrar</span>
    <h1 class="mt-3">Acessar área de gestão</h1>

    <form v-if="step === 'email'" @submit.prevent="sendCode" class="shell mt-6">
      <div class="card !mb-0 flex flex-col gap-3">
        <label class="field">
          E-mail
          <input v-model="email" type="email" required class="input" />
        </label>
        <button type="submit" class="btn btn-primary self-start">Enviar código</button>
      </div>
    </form>

    <form v-else @submit.prevent="submitCode" class="shell mt-6">
      <div class="card !mb-0 flex flex-col gap-3">
        <p class="text-sm text-neutral-500 dark:text-neutral-400">
          Enviamos um código para {{ email }}.
        </p>
        <label class="field">
          Código
          <input v-model="code" required class="input" />
        </label>
        <p v-if="codeError" class="text-sm text-red-600 dark:text-red-400">{{ codeError }}</p>
        <button type="submit" class="btn btn-primary self-start">Entrar</button>
      </div>
    </form>
  </section>
</template>
