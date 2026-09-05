<script setup lang="ts">
import { computed } from 'vue'
import { displayMessage } from '../api/errorMessages'
import type { ProblemDetails } from '../types/api'

const props = defineProps<{ problem: ProblemDetails | null }>()

const message = computed(() => (props.problem ? displayMessage(props.problem) : ''))
</script>

<template>
  <div
    v-if="problem"
    class="my-2 rounded-md border border-red-300 bg-red-50 px-4 py-3 text-left text-red-800"
  >
    <p>{{ message }}</p>
    <ul v-if="problem.errors?.length" class="mt-1.5 list-disc pl-5">
      <li v-for="fieldError in problem.errors" :key="fieldError.field">
        {{ fieldError.field }}: {{ fieldError.detail }}
      </li>
    </ul>
  </div>
</template>
