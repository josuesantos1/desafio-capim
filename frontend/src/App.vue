<script setup lang="ts"></script>

<template>
  <header class="border-b border-slate-200 px-6 py-4 text-left dark:border-slate-700">
    <router-link to="/" class="font-semibold text-slate-900 no-underline dark:text-slate-100">
      clinic-api
    </router-link>
  </header>
  <main class="p-6 text-left">
    <router-view />
  </main>
</template>
