<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useDentistsStore } from '../stores/dentists'
import ErrorBanner from './ErrorBanner.vue'
import type { Dentist, DentistCreateInput, DentistUpdateInput } from '../types/api'

const props = defineProps<{ clinicId: string }>()

const store = useDentistsStore()

const form = reactive<DentistCreateInput>({ name: '', phone: '', email: '' })
const bioText = ref('')
const specialtiesText = ref('')
const yearsText = ref('')

const editingId = ref<string | null>(null)
const editForm = reactive<DentistUpdateInput>({ name: '', phone: '', email: '' })
const editBioText = ref('')
const editSpecialtiesText = ref('')
const editYearsText = ref('')

onMounted(() => {
  store.list(props.clinicId).catch(() => {
    // erro já está em store.error
  })
})

function normalizeSpecialties(text: string): string[] {
  return [...new Set(text.split('\n').map((s) => s.trim()).filter((s) => s !== ''))]
}

async function submit() {
  try {
    await store.create(props.clinicId, {
      ...form,
      bio: bioText.value,
      specialties: normalizeSpecialties(specialtiesText.value),
      years_of_experience: Number(yearsText.value) || 0,
    })
    form.name = ''
    form.phone = ''
    form.email = ''
    bioText.value = ''
    specialtiesText.value = ''
    yearsText.value = ''
  } catch {
    // erro já está em store.error
  }
}

function startEdit(dentist: Dentist) {
  editingId.value = dentist.id
  editForm.name = dentist.name
  editForm.phone = dentist.phone
  editForm.email = dentist.email
  editBioText.value = dentist.bio
  editSpecialtiesText.value = dentist.specialties.join('\n')
  editYearsText.value = String(dentist.years_of_experience)
}

function cancelEdit() {
  editingId.value = null
}

async function submitEdit(dentistId: string) {
  try {
    await store.update(props.clinicId, dentistId, {
      ...editForm,
      bio: editBioText.value,
      specialties: normalizeSpecialties(editSpecialtiesText.value),
      years_of_experience: Number(editYearsText.value) || 0,
    })
    editingId.value = null
  } catch {
    // erro já está em store.error
  }
}

async function toggleRole(dentistId: string, role: 'is_administrator' | 'is_legal_representative', value: boolean) {
  try {
    await store.updateRoles(props.clinicId, dentistId, { [role]: value })
  } catch {
    // store.error já reflete o motivo (ex. LAST_ADMIN_REQUIRED); o checkbox
    // volta ao valor original porque a lista é re-renderizada a partir do store
  }
}

async function remove(dentistId: string) {
  if (!confirm('Excluir este dentista?')) return
  try {
    await store.remove(props.clinicId, dentistId)
  } catch {
    // erro já está em store.error
  }
}
</script>

<template>
  <div>
    <ErrorBanner :problem="store.error" />

    <div class="shell max-w-md">
      <form @submit.prevent="submit" class="card !mb-0 flex flex-col gap-3">
        <h2 class="mt-0">Novo dentista</h2>
        <label class="field">
          Nome
          <input v-model="form.name" required class="input" />
        </label>
        <label class="field">
          Telefone
          <input v-model="form.phone" required class="input" />
        </label>
        <label class="field">
          E-mail
          <input v-model="form.email" type="email" required class="input" />
        </label>
        <label class="field">
          Bio
          <textarea v-model="bioText" rows="2" class="input" />
        </label>
        <label class="field">
          Especialidades (uma por linha)
          <textarea v-model="specialtiesText" rows="2" class="input" />
        </label>
        <label class="field">
          Anos de experiência
          <input v-model="yearsText" type="number" min="0" class="input" />
        </label>
        <button type="submit" :disabled="store.loading" class="btn btn-primary self-start">
          Adicionar dentista
        </button>
      </form>
    </div>

    <h2>Dentistas</h2>
    <p v-if="store.listByClinic(clinicId).length === 0" class="text-neutral-500 dark:text-neutral-400">
      Nenhum dentista cadastrado.
    </p>
    <ul v-else class="list-plain">
      <li v-for="dentist in store.listByClinic(clinicId)" :key="dentist.id" class="list-item">
        <template v-if="editingId === dentist.id">
          <form @submit.prevent="submitEdit(dentist.id)" class="flex flex-col gap-2">
            <div class="flex flex-wrap items-center gap-2">
              <input v-model="editForm.name" required class="input" />
              <input v-model="editForm.phone" required class="input" />
              <input v-model="editForm.email" type="email" required class="input" />
            </div>
            <textarea v-model="editBioText" rows="2" class="input" placeholder="Bio" />
            <textarea
              v-model="editSpecialtiesText"
              rows="2"
              class="input"
              placeholder="Especialidades (uma por linha)"
            />
            <input v-model="editYearsText" type="number" min="0" class="input" placeholder="Anos de experiência" />
            <div class="flex gap-2">
              <button type="submit" :disabled="store.loading" class="btn">Salvar</button>
              <button type="button" class="btn" @click="cancelEdit">Cancelar</button>
            </div>
          </form>
        </template>
        <template v-else>
          <div class="flex flex-wrap items-center gap-3">
            <strong>{{ dentist.name }}</strong>
            <span>{{ dentist.email }} — {{ dentist.phone }}</span>
            <label class="field-inline">
              <input
                type="checkbox"
                :checked="dentist.is_administrator"
                @change="toggleRole(dentist.id, 'is_administrator', ($event.target as HTMLInputElement).checked)"
              />
              Administrador
            </label>
            <label class="field-inline">
              <input
                type="checkbox"
                :checked="dentist.is_legal_representative"
                @change="toggleRole(dentist.id, 'is_legal_representative', ($event.target as HTMLInputElement).checked)"
              />
              Representante legal
            </label>
            <button type="button" class="btn" @click="startEdit(dentist)">Editar</button>
            <button type="button" class="btn" @click="remove(dentist.id)">Excluir</button>
          </div>
        </template>
      </li>
    </ul>
  </div>
</template>
