<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useClinicsStore } from '../stores/clinics'
import { mockClinicProfile } from '../mock/profile'

const router = useRouter()
const clinicsStore = useClinicsStore()

const query = ref('')
const location = ref('')

function submit() {
  router.push({ path: '/search', query: { q: query.value } })
}
</script>

<template>
  <section>
    <h1>Encontre uma clínica odontológica</h1>

    <form @submit.prevent="submit" class="form">
      <label class="field">
        Buscar clínica, dentista ou especialidade
        <input v-model="query" class="input" />
      </label>
      <label class="field">
        Localização
        <input v-model="location" placeholder="Cidade (opcional)" class="input" />
      </label>
      <button type="submit" class="btn">Buscar</button>
    </form>

    <h2>Clínicas em destaque</h2>
    <p v-if="clinicsStore.list.length === 0">
      Nenhuma clínica disponível ainda —
      <router-link to="/clinics" class="hover:underline">crie uma pela área de gestão</router-link
      >.
    </p>
    <ul v-else class="list-plain">
      <li v-for="clinic in clinicsStore.list" :key="clinic.id" class="list-item">
        <router-link :to="`/c/${clinic.id}`" class="hover:underline">
          <strong>{{ clinic.trade_name }}</strong>
          <span v-if="clinic.status === 'pending'" class="badge">Em configuração</span>
          <br />
          ★ {{ mockClinicProfile(clinic).rating }} — {{ mockClinicProfile(clinic).city }}
        </router-link>
      </li>
    </ul>
  </section>
</template>
