<script setup lang="ts">
import { computed } from 'vue'
import { displayMessage } from '../api/errorMessages'
import type { ProblemDetails } from '../types/api'

const props = defineProps<{ problem: ProblemDetails | null }>()

const message = computed(() => (props.problem ? displayMessage(props.problem) : ''))
</script>

<template>
  <div v-if="problem" class="error-banner">
    <p>{{ message }}</p>
    <ul v-if="problem.errors?.length">
      <li v-for="fieldError in problem.errors" :key="fieldError.field">
        {{ fieldError.field }}: {{ fieldError.detail }}
      </li>
    </ul>
  </div>
</template>

<style scoped>
.error-banner {
  background: #fdecec;
  border: 1px solid #f5b5b5;
  color: #8a1f1f;
  border-radius: 6px;
  padding: 10px 14px;
  margin: 8px 0;
  text-align: left;
}

.error-banner ul {
  margin: 6px 0 0;
  padding-left: 18px;
}
</style>
