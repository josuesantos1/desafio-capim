<script setup lang="ts"></script>

<template>
  <header>
    <router-link to="/clinics">clinic-api</router-link>
  </header>
  <main>
    <router-view />
  </main>
</template>

<style scoped>
header {
  padding: 16px 24px;
  border-bottom: 1px solid var(--border);
  text-align: left;
}

header a {
  color: var(--text-h);
  text-decoration: none;
  font-weight: 600;
}

main {
  padding: 24px;
  text-align: left;
}
</style>
