<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'

const router = useRouter()
const authStore = useAuthStore()

function logout() {
  authStore.logout()
  router.push('/')
}
</script>

<template>
  <div class="fixed inset-0 -z-10 bg-neutral-50 dark:bg-neutral-950" />

  <nav class="sticky top-6 z-20 mx-auto mt-6 w-max">
    <div
      class="flex items-center gap-6 rounded-full border border-neutral-900/10 bg-white/70 px-5 py-2.5 shadow-[0_1px_2px_rgba(15,15,15,0.04),0_16px_40px_-16px_rgba(15,15,15,0.2)] backdrop-blur-2xl dark:border-white/10 dark:bg-neutral-900/70"
    >
      <router-link
        to="/"
        class="font-display text-sm font-semibold text-neutral-900 no-underline dark:text-neutral-50"
      >
        Capim Hub
      </router-link>
      <router-link
        to="/search"
        class="text-sm text-neutral-500 no-underline transition-colors hover:text-neutral-900 dark:hover:text-neutral-100"
      >
        Buscar
      </router-link>
      <router-link
        to="/clinics"
        class="text-sm text-neutral-500 no-underline transition-colors hover:text-neutral-900 dark:hover:text-neutral-100"
      >
        Área de gestão
      </router-link>
      <template v-if="authStore.email">
        <span class="text-sm text-neutral-500 dark:text-neutral-400">{{ authStore.email }}</span>
        <button
          type="button"
          class="cursor-pointer border-0 bg-transparent p-0 text-sm text-neutral-500 transition-colors hover:text-neutral-900 dark:hover:text-neutral-100"
          @click="logout"
        >
          Sair
        </button>
      </template>
      <router-link
        v-else
        to="/login"
        class="text-sm text-neutral-500 no-underline transition-colors hover:text-neutral-900 dark:hover:text-neutral-100"
      >
        Entrar
      </router-link>
    </div>
  </nav>

  <main class="mx-auto max-w-5xl px-4 py-16 text-left sm:px-6">
    <router-view />
  </main>
</template>
