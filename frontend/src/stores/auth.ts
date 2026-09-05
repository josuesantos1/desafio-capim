import { defineStore } from 'pinia'
import { ref } from 'vue'

const STORAGE_KEY = 'auth.email'

export const useAuthStore = defineStore('auth', () => {
  const email = ref<string | null>(localStorage.getItem(STORAGE_KEY))

  function login(newEmail: string) {
    email.value = newEmail
    localStorage.setItem(STORAGE_KEY, newEmail)
  }

  function logout() {
    email.value = null
    localStorage.removeItem(STORAGE_KEY)
  }

  return { email, login, logout }
})
